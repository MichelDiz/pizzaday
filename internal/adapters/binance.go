package adapters

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/MichelDiz/pizzaday/internal/helpers"
	"github.com/MichelDiz/pizzaday/internal/models"
)

type BinanceAdapter struct {
	Symbol               string
	Streams              []string
	AggregatedBuyTrades  []models.Trade
	AggregatedSellTrades []models.Trade
	MinTradeSize         float64
	LastTradePrice       float64

	// Estatísticas
	BuyCount   int
	SellCount  int
	TotalCount int

	BuyVolume  float64
	SellVolume float64

	TotalFreq float64
	BuyFreq   float64
	SellFreq  float64

	BuyActivity  float64
	SellActivity float64

	KdBuy  float64
	KdSell float64

	LastUpdate time.Time
}

func (b BinanceAdapter) GetURL() string {
	streamsQuery := strings.Join(b.Streams, "/")
	return "wss://fstream.binance.com/stream?streams=" + streamsQuery
}

type CombinedMessage struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

func (b *BinanceAdapter) ProcessMessage(messageType int, message []byte, verbose bool) {
	if verbose {
		log.Printf("Message received: Type: %d, Size: %d bytes", messageType, len(message))
	}

	b.MinTradeSize = 0.11

	var combined CombinedMessage
	if err := json.Unmarshal(message, &combined); err != nil {
		log.Println("Error decoding combined message:", err)
		return
	}

	if verbose {
		log.Printf("Stream: %s", combined.Stream)
	}

	switch combined.Stream {
	case b.Symbol + "@trade":
		var newTrade models.Trade
		if err := json.Unmarshal(combined.Data, &newTrade); err != nil {
			log.Println("Error decoding trade:", err)
			return
		}

		newQty, err := strconv.ParseFloat(newTrade.Quantity, 64)
		if err != nil {
			log.Printf("Error converting Quantity from new trade: %v", err)
			return
		}

		newPrice, err := strconv.ParseFloat(newTrade.Price, 64)
		if err == nil {
			b.LastTradePrice = newPrice
		}

		b.TotalCount++

		if newTrade.Buyer {
			b.SellCount++
			b.SellVolume += newQty
			b.AggregatedSellTrades = append(b.AggregatedSellTrades, newTrade)
		} else {
			b.BuyCount++
			b.BuyVolume += newQty
			b.AggregatedBuyTrades = append(b.AggregatedBuyTrades, newTrade)
		}

		now := time.Now()

		duration := now.Sub(b.LastUpdate).Minutes()
		if duration < (1.0 / 60.0) { // Se menor que 1 segundo, trata como 1 segundo
			duration = 1.0 / 60.0
		}

		if duration > 0 {
			b.TotalFreq = float64(b.TotalCount) / 60
			b.BuyFreq = float64(b.BuyCount) / 60
			b.SellFreq = float64(b.SellCount) / 60
		}

		// Calcular atividade relativa de compra e venda
		b.BuyActivity = float64(b.BuyCount) / (b.BuyVolume + 1) // Evita divisão por zero
		b.SellActivity = float64(b.SellCount) / (b.SellVolume + 1)

		// Calcular relação de compra e venda (KD)
		b.KdBuy = helpers.CalculateKD(float64(b.BuyCount), float64(b.SellCount))
		b.KdSell = helpers.CalculateKD(float64(b.SellCount), float64(b.BuyCount))

		if now.Sub(b.LastUpdate) > 5*time.Second {
			b.LastUpdate = now
		}

		if verbose {
			log.Printf("Stats: Total=%d | BuyCount=%d (Vol: %.4f) | SellCount=%d (Vol: %.4f)",
				b.TotalCount, b.BuyCount, b.BuyVolume, b.SellCount, b.SellVolume)

			log.Printf("Frequências: Total=%.2f/min | Buy=%.2f/min | Sell=%.2f/min",
				b.TotalFreq, b.BuyFreq, b.SellFreq)

			log.Printf("Atividade: Buy=%.4f | Sell=%.4f | KdBuy=%.2f | KdSell=%.2f",
				b.BuyActivity, b.SellActivity, b.KdBuy, b.KdSell)
		}

		var targetList *[]models.Trade
		if newTrade.Buyer {
			targetList = &b.AggregatedSellTrades
		} else {
			targetList = &b.AggregatedBuyTrades
		}

		if len(*targetList) > 0 {
			lastAgg := &(*targetList)[0]
			aggQty, err := strconv.ParseFloat(lastAgg.Quantity, 64)
			if err != nil {
				log.Printf("Error converting Quantity from aggregated trade: %v", err)
			}

			if lastAgg.Time == newTrade.Time && lastAgg.Buyer == newTrade.Buyer {
				aggQty += newQty
				lastAgg.Quantity = fmt.Sprintf("%.1f", aggQty)
				lastAgg.Price = newTrade.Price
			} else if newQty >= b.MinTradeSize {
				*targetList = append([]models.Trade{newTrade}, *targetList...)
				log.Printf("Beep: new aggregated %s trade with Quantity %s BTC",
					map[bool]string{true: "SELL", false: "BUY"}[newTrade.Buyer], newTrade.Quantity)
				log.Printf("%s", helpers.ConvertBtcToUsd(newTrade.Quantity, b.LastTradePrice))
			}
		} else if newQty >= b.MinTradeSize {
			*targetList = append(*targetList, newTrade)
			log.Printf("Beep: aggregated %s trade with Quantity %s BTC",
				map[bool]string{true: "SELL", false: "BUY"}[newTrade.Buyer], newTrade.Quantity)
			log.Printf("%s", helpers.ConvertBtcToUsd(newTrade.Quantity, b.LastTradePrice))
		}

	case b.Symbol + "@bookTicker":
		var orderBook models.OrderBook
		if err := json.Unmarshal(combined.Data, &orderBook); err != nil {
			log.Println("Error decoding orderBook:", err)
		} else {
			fmt.Printf("[OrderBook] BID: %s (%s) | ASK: %s (%s)\n", orderBook.BidPrice, orderBook.BidQty, orderBook.AskPrice, orderBook.AskQty)
		}

	case b.Symbol + "@forceOrder":
		var liquidation models.Liquidation
		if err := json.Unmarshal(combined.Data, &liquidation); err != nil {
			log.Println("Error decoding liquidation:", err)
		} else {
			fmt.Printf("[Liquidation] Price: %s, Quantity: %s, Side: %s\n",
				liquidation.Order.Price, liquidation.Order.Quantity, liquidation.Order.Side)
		}

	default:
		log.Printf("Unhandled stream: %s", combined.Stream)
	}
}
