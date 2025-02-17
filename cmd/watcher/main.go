package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/MichelDiz/pizzaday/internal/adapters"
	"github.com/MichelDiz/pizzaday/internal/helpers"
)

func main() {

	verbose := helpers.VerboseLog()

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

	adapter := adapters.BinanceAdapter{
		Symbol:  symbol,
		Streams: streams,
	}

	// Start WebSocket connection with selected streams
	helpers.Connect(&adapter, verbose)
}
