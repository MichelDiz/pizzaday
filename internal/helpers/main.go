package helpers

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var verbose bool

func init() {
	verbose = os.Getenv("VERBOSE") == "true"
}

func VerboseLog() bool {
	return verbose
}

type Adapter interface {
	GetURL() string
	ProcessMessage(messageType int, message []byte, verbose bool)
}

func Connect(adapter Adapter, verbose bool) {
	u, err := url.Parse(adapter.GetURL())
	if err != nil {
		log.Fatal("Invalid URL:", err)
	}

	fmt.Printf("Connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer c.Close()

	log.Println("Connection successfully established.")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			adapter.ProcessMessage(messageType, message, verbose)
		}
	}()

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
