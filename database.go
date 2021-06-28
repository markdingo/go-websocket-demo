package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Ticker struct {
	Symbol     string
	UpdateTime time.Time
	Price      float64 // Cents
}

func (t *Ticker) String() string {
	return fmt.Sprintf("%s: $%0.2f", t.Symbol, t.Price)
}

// Database is a mock implementation of a stock exchange feed. It contains Tickers which
// are simply symbols with prices.
type Database struct {
	mu   sync.Mutex // Protects map
	data map[string]*Ticker
}

// Add adds the symbol into the database if its not already present
func (t *Database) Add(symbol string, price float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.data[symbol]; !ok {
		t.data[symbol] = &Ticker{Symbol: symbol, Price: price}
	}
}

// randomUpdater - as its name implies - randomly updates tickers in the database. The
// Price should generally oscillate around its original value.
func randomUpdater(DB *Database) (changed []*Ticker) {
	DB.mu.Lock()
	defer DB.mu.Unlock()

	every := rand.Intn(len(DB.data))
	if every < 1 {
		every = 1
	}

	ix := 0
	for _, ticker := range DB.data {
		ix++
		if ix < every {
			continue
		}
		ix = 0
		ticker.UpdateTime = time.Now()
		ticker.Price = 1.0 + ticker.Price*(rand.Float64()+0.5)
		changed = append(changed, ticker)
		log.Println(ticker.String())
	}

	return
}
