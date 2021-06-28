package main

import (
	"context"
	"log"
	"sync"

	"nhooyr.io/websocket"
)

// Notifier keeps track of all clients and the tickers they're interested in. Notifier is
// responsible for sending out a TickerChange message to every Client which has registered
// interest in the ticker in question.
type Notifier struct {
	mu      sync.Mutex // Protects map
	clients map[*Client]struct{}
}

func (t *Notifier) Add(c *Client) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.clients[c] = struct{}{}
	log.Println("Client Added", c.id, "Total", len(t.clients))
}

func (t *Notifier) Delete(c *Client) {
	if c == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.clients, c)
	log.Println("Client Deleted", c.id, "Total", len(t.clients))
}

type Client struct {
	mu         sync.Mutex
	id         string
	generation int
	tickers    []string
	conn       *websocket.Conn
}

// Notify sends a TickerChange message to every client who has registered interest in any
// of the changed tickers. Each client is sent a single message with multiple tickers. A
// go routine is spun up to send the message to each client to ensure the notifier doesn't
// stall.
func (t *Notifier) Notify(DB *Database, changed []*Ticker) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Match up all client tickers with the changed slice.
	for client := range t.clients {
		tickers := make([]Ticker, 0)
		for _, changedTicker := range changed {
			for _, clientTicker := range client.tickers {
				if changedTicker.Symbol == clientTicker {
					tickers = append(tickers, *changedTicker) // Copy is important
				}
			}
		}

		// If any tickers match, start up a go routine to send changes.
		if len(tickers) > 0 {
			go client.Notify(tickers)
		}
	}
}

// Notify sends a TickerChange message to the client. A copy of the Tickers is passed to
// this function as the original Ticker from the database is no longer protected by the
// database mutex.
func (t *Client) Notify(tickers []Ticker) {
	msg := &TickerChange{Changes: make([]SymbolChange, 0, len(tickers))}
	for _, ticker := range tickers {
		msg.Changes = append(msg.Changes,
			SymbolChange{Symbol: ticker.Symbol, UpdateTime: ticker.UpdateTime.Unix(),
				Price: ticker.Price})
	}
	WriteMessage(context.Background(), t.conn, msg)
}
