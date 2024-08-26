package ohlc

import (
	"sync"
	"time"
)

type Candlestick struct {
	Symbol    string    // The trading symbol (e.g., BTCUSDT)
	Timestamp time.Time // The timestamp for the candlestick
	Open      float64   // The opening price
	High      float64   // The highest price during the timeframe
	Low       float64   // The lowest price during the timeframe
	Close     float64   // The closing price
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
func (a *Aggregator) ProcessTick(symbol string, price float64, timestamp int64) (Candlestick, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	t := time.Unix(0, timestamp*int64(time.Millisecond)).Truncate(time.Minute)
	cs, exists := a.candlesticks[symbol]

	var completedCandlestick Candlestick
	isComplete := false

	if !exists || t.After(cs.Timestamp) {
		if exists {
			// The previous candlestick is complete, so broadcast it
			completedCandlestick = cs
			isComplete = true
			a.broadcastCandlestick(completedCandlestick)
		}
		// Create a new candlestick with the symbol
		cs = Candlestick{
			Symbol:    symbol,
			Timestamp: t,
			Open:      price,
			High:      price,
			Low:       price,
			Close:     price,
		}
	} else {
		// Update the existing candlestick
		if price > cs.High {
			cs.High = price
		}
		if price < cs.Low {
			cs.Low = price
		}
		cs.Close = price
	}

	a.candlesticks[symbol] = cs
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
