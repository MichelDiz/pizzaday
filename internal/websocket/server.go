package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/MichelDiz/pizzaday/internal/adapters"
	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	upgrader    websocket.Upgrader
	connections map[*websocket.Conn]bool
	mu          sync.Mutex
	adapter     *adapters.BinanceAdapter
}

func NewWebSocketServer(adapter *adapters.BinanceAdapter) *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections: make(map[*websocket.Conn]bool),
		adapter:     adapter,
	}
}

func (s *WebSocketServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.connections[conn] = true
	s.mu.Unlock()

	log.Println("New client connected")

	for {
		if _, _, err := conn.NextReader(); err != nil {
			s.mu.Lock()
			delete(s.connections, conn)
			s.mu.Unlock()
			log.Println("Client disconnected")
			break
		}
	}
}

func (s *WebSocketServer) broadcast() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		adapterSnapshot := struct {
			Symbol         string    `json:"Symbol"`
			Streams        []string  `json:"Streams"`
			MinTradeSize   float64   `json:"MinTradeSize"`
			LastTradePrice float64   `json:"LastTradePrice"`
			BuyCount       int       `json:"BuyCount"`
			SellCount      int       `json:"SellCount"`
			TotalCount     int       `json:"TotalCount"`
			BuyVolume      float64   `json:"BuyVolume"`
			SellVolume     float64   `json:"SellVolume"`
			TotalFreq      float64   `json:"TotalFreq"`
			BuyFreq        float64   `json:"BuyFreq"`
			SellFreq       float64   `json:"SellFreq"`
			BuyActivity    float64   `json:"BuyActivity"`
			SellActivity   float64   `json:"SellActivity"`
			KdBuy          float64   `json:"KdBuy"`
			KdSell         float64   `json:"KdSell"`
			LastUpdate     time.Time `json:"LastUpdate"`
		}{
			Symbol:         s.adapter.Symbol,
			Streams:        s.adapter.Streams,
			MinTradeSize:   s.adapter.MinTradeSize,
			LastTradePrice: s.adapter.LastTradePrice,
			BuyCount:       s.adapter.BuyCount,
			SellCount:      s.adapter.SellCount,
			TotalCount:     s.adapter.TotalCount,
			BuyVolume:      s.adapter.BuyVolume,
			SellVolume:     s.adapter.SellVolume,
			TotalFreq:      s.adapter.TotalFreq,
			BuyFreq:        s.adapter.BuyFreq,
			SellFreq:       s.adapter.SellFreq,
			BuyActivity:    s.adapter.BuyActivity,
			SellActivity:   s.adapter.SellActivity,
			KdBuy:          s.adapter.KdBuy,
			KdSell:         s.adapter.KdSell,
			LastUpdate:     s.adapter.LastUpdate,
		}
		s.mu.Unlock()

		adapterData, err := json.Marshal(adapterSnapshot)
		if err != nil {
			log.Println("Error serializing BinanceAdapter data:", err)
			continue
		}

		s.mu.Lock()
		for conn := range s.connections {
			if err := conn.WriteMessage(websocket.TextMessage, adapterData); err != nil {
				log.Println("Error sending WebSocket message:", err)
				conn.Close()
				delete(s.connections, conn)
			}
		}
		s.mu.Unlock()
	}
}

func (s *WebSocketServer) Start() {
	http.HandleFunc("/ws", s.handleConnection)

	go s.broadcast()

	log.Println("WebSocket server started on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("WebSocket server error:", err)
	}
}
