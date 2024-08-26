package binance

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Tick struct {
	Symbol    string  `json:"s"`
	Price     float64 `json:"p,string"`
	Timestamp int64   `json:"E"`
}

func Connect(symbols []string, tickChan chan<- Tick) {
	wsURL := "wss://stream.binance.com:9443/ws"
	for _, symbol := range symbols {
		wsURL += "/" + symbol + "@aggTrade"
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			continue
		}

		var tick Tick
		err = json.Unmarshal(message, &tick)
		if err != nil {
			log.Println("Error unmarshalling tick data:", err)
			continue
		}

		tickChan <- tick
	}
}
