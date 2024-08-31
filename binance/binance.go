package binance

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Define the Tick struct to match the data inside the "data" field
type Tick struct {
	EventType     string  `json:"e"`        // Event type
	Timestamp     int64   `json:"E"`        // Event timestamp
	Symbol        string  `json:"s"`        // Symbol
	AggregateID   int64   `json:"a"`        // Aggregate trade ID
	Price         float64 `json:"p,string"` // Price as a string in JSON, converted to float64
	Quantity      float64 `json:"q,string"` // Quantity as a string in JSON, converted to float64
	FirstTradeID  int64   `json:"f"`        // First trade ID
	LastTradeID   int64   `json:"l"`        // Last trade ID
	TradeTime     int64   `json:"T"`        // Trade time
	IsMarketMaker bool    `json:"m"`        // Market maker indicator
	Ignore        bool    `json:"M"`        // Ignore field
}

// Define a struct to capture the outer layer of the JSON
type StreamData struct {
	Stream string `json:"stream"`
	Data   Tick   `json:"data"` // Nested Tick data
}

// Connect establishes a WebSocket connection to Binance and sends Tick data to the provided channel
func Connect(symbols []string, tickChan chan<- Tick) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		wsURL := "wss://stream.binance.com:9443/stream?streams="
		streams := []string{}
		for _, symbol := range symbols {
			streams = append(streams, strings.ToLower(symbol)+"@aggTrade")
		}
		wsURL += strings.Join(streams, "/") // Join symbols with the correct WebSocket URL format

		log.Printf("Connecting to WebSocket URL: %s\n", wsURL) // Log the WebSocket URL for verification

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			log.Println("Error connecting to WebSocket:", err)
			time.Sleep(10 * time.Second) // Wait before retrying
			continue
		}

		go func() {
			defer conn.Close()
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("Error reading WebSocket message:", err)
					return // Exit the goroutine to reconnect
				}
				log.Println("Message:", string(message))

				var streamData StreamData
				err = json.Unmarshal(message, &streamData)
				if err != nil {
					log.Println("Error unmarshalling stream data:", err)
					log.Println("Message:", string(message))
					continue
				}

				tick := streamData.Data
				tickChan <- tick
			}
		}()

		// Wait for the ticker to trigger a new connection
		<-ticker.C
	}
}
