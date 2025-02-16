package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/MichelDiz/pizzaday/internal/models"
	"github.com/gorilla/websocket"
)

type CombinedMessage struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

func connectWebSocket(symbol string, streams []string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	streamsQuery := strings.Join(streams, "/")

	u := url.URL{
		Scheme:   "wss",
		Host:     "fstream.binance.com",
		Path:     "/stream",
		RawQuery: "streams=" + streamsQuery,
	}

	fmt.Printf("Connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer c.Close()

	log.Println("Connection successfully established.")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			msgType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			log.Printf("Message received: Type: %d, Size: %d bytes, Content: %s", msgType, len(message), message)

			// Decode the message as a combined message
			var combined CombinedMessage
			if err := json.Unmarshal(message, &combined); err != nil {
				log.Println("Error decoding combined message:", err)
				continue
			}

			log.Printf("Stream: %s", combined.Stream)

			// Check the type of stream received
			switch combined.Stream {
			case symbol + "@trade":
				var trade models.Trade
				if err := json.Unmarshal(combined.Data, &trade); err != nil {
					log.Println("Error decoding trade:", err)
				} else {
					fmt.Printf("[Trade] Price: %s, Quantity: %s, Buyer? %v\n", trade.Price, trade.Quantity, !trade.Buyer)
				}
			case symbol + "@bookTicker":
				var orderBook models.OrderBook
				if err := json.Unmarshal(combined.Data, &orderBook); err != nil {
					log.Println("Error decoding orderBook:", err)
				} else {
					fmt.Printf("[OrderBook] BID: %s (%s) | ASK: %s (%s)\n", orderBook.BidPrice, orderBook.BidQty, orderBook.AskPrice, orderBook.AskQty)
				}
			case symbol + "@forceOrder":
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
	}()

	// Keep the connection alive until interrupted
	for {
		select {
		case <-interrupt:
			fmt.Println("Interrupting connection...")
			_ = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			time.Sleep(time.Second)
			return
		}
	}
}

func main() {
	// Flags to enable specific streams
	enableTrade := flag.Bool("trade", false, "Enable trade stream")
	enableOrderBook := flag.Bool("orderbook", false, "Enable book ticker stream")
	enableLiquidation := flag.Bool("forceorder", false, "Enable force order stream")

	// Flag to exclude streams (e.g., -exclude=trade,orderbook)
	excludeStreams := flag.String("exclude", "", "Streams to be excluded, separated by commas (e.g., trade,orderbook)")

	flag.Parse()

	symbol := "btcusdt"

	// Define available streams
	defaultStreams := map[string]bool{
		"trade":      true,
		"orderbook":  true,
		"forceorder": true,
	}

	// If specific flags are passed, disable unspecified streams
	if *enableTrade || *enableOrderBook || *enableLiquidation {
		defaultStreams = map[string]bool{
			"trade":      *enableTrade,
			"orderbook":  *enableOrderBook,
			"forceorder": *enableLiquidation,
		}
	}

	// If the `-exclude` flag is used, remove the specified streams
	if *excludeStreams != "" {
		excludes := strings.Split(*excludeStreams, ",")
		for _, stream := range excludes {
			defaultStreams[strings.TrimSpace(stream)] = false
		}
	}

	// Filter active streams
	var streams []string
	for stream, active := range defaultStreams {
		if active {
			streams = append(streams, fmt.Sprintf("%s@%s", symbol, stream))
		}
	}

	if len(streams) == 0 {
		fmt.Println("No stream enabled. Exiting.")
		return
	}

	// Start WebSocket connection with selected streams
	connectWebSocket(symbol, streams)
}
