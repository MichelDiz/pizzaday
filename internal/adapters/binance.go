package adapters

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/MichelDiz/pizzaday/internal/models"
)

type BinanceAdapter struct {
	Symbol  string
	Streams []string
}

func (b BinanceAdapter) GetURL() string {
	streamsQuery := strings.Join(b.Streams, "/")
	return "wss://fstream.binance.com/stream?streams=" + streamsQuery
}

type CombinedMessage struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

// ProcessMessage decodifica e processa as mensagens recebidas.
func (b BinanceAdapter) ProcessMessage(messageType int, message []byte) {
	log.Printf("Message received: Type: %d, Size: %d bytes", messageType, len(message))

	var combined CombinedMessage
	if err := json.Unmarshal(message, &combined); err != nil {
		log.Println("Error decoding combined message:", err)
		return
	}
	log.Printf("Stream: %s", combined.Stream)

	switch combined.Stream {
	case b.Symbol + "@trade":
		var trade models.Trade
		if err := json.Unmarshal(combined.Data, &trade); err != nil {
			log.Println("Error decoding trade:", err)
		} else {
			fmt.Printf("[Trade] Price: %s, Quantity: %s, Buyer? %v\n", trade.Price, trade.Quantity, !trade.Buyer)
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
