package ohlc

import (
	"sync"
	"time"
	"trading-chart-service/binance"
)

type Candlestick struct {
	Symbol    string
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
}

type Aggregator struct {
	mu           sync.Mutex
	candlesticks map[string]Candlestick
	subscribers  []chan Candlestick
}

// NewAggregator creates a new Aggregator instance
func NewAggregator() *Aggregator {
	return &Aggregator{
		candlesticks: make(map[string]Candlestick),
		subscribers:  []chan Candlestick{},
	}
}

// ProcessTick processes a new tick, updates the current candlestick, and returns the completed candlestick if it is done
func (a *Aggregator) ProcessTick(tick binance.Tick) (Candlestick, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Convert timestamp to time.Time
	timestamp := time.Unix(0, tick.Timestamp*int64(time.Millisecond)).Truncate(time.Minute)
	cs, exists := a.candlesticks[tick.Symbol]

	var completedCandlestick Candlestick
	isComplete := false

	if !exists || timestamp.After(cs.Timestamp) {
		if exists {
			// The previous candlestick is complete, so broadcast it
			completedCandlestick = cs
			isComplete = true
			a.broadcastCandlestick(completedCandlestick)
		}
		// Create a new candlestick with the symbol
		cs = Candlestick{
			Symbol:    tick.Symbol,
			Timestamp: timestamp,
			Open:      tick.Price,
			High:      tick.Price,
			Low:       tick.Price,
			Close:     tick.Price,
		}
	} else {
		// Update the existing candlestick
		if tick.Price > cs.High {
			cs.High = tick.Price
		}
		if tick.Price < cs.Low {
			cs.Low = tick.Price
		}
		cs.Close = tick.Price
	}

	a.candlesticks[tick.Symbol] = cs
	return completedCandlestick, isComplete
}

// Subscribe allows clients to receive completed candlesticks
func (a *Aggregator) Subscribe() chan Candlestick {
	ch := make(chan Candlestick)
	a.mu.Lock()
	a.subscribers = append(a.subscribers, ch)
	a.mu.Unlock()
	return ch
}

// broadcastCandlestick broadcasts the completed candlestick to all subscribers
func (a *Aggregator) broadcastCandlestick(candlestick Candlestick) {
	for _, ch := range a.subscribers {
		select {
		case ch <- candlestick:
		default:
			// If the channel is full, skip this subscriber
		}
	}
}
