package main

// This demo server accepts websocket connections and takes requests to watch a set of
// ticker names. Whenever any of the tickers change, a message is sent back to the client
// with the change details.
//
// The ticker database is similated. It is pre-populated here and also includes any news
// symbols that the client requests. It is periodically and randomly updated by a
// background go-routine.
//
// The server responds to application-level ping messages.
//
// The most salient feature is that the server notices when clients become unresponsive
// due to a lack of ping messages. Similarly, the client notices when the server becomes
// unresponsive due to a lack of pong response. Note that these are application level ping
// and pong messages, not the websocket ones because pong is not exposed in the websocket
// package.

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"nhooyr.io/websocket"
)

const (
	Name = "server"
)

func main() {
	log.SetPrefix(Name + " ")
	log.SetFlags(log.Ltime)

	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage:", Name, "ListenAddress:ListenPort")
		return
	}

	// Create the global data structures
	NF := &Notifier{clients: make(map[*Client]struct{})}

	DB := &Database{data: make(map[string]*Ticker)}
	DB.Add("CMG", 1518.75) // Pre-seed with something
	DB.Add("AAPL", 133.11)
	DB.Add("SPY", 426.61)
	DB.Add("MOAT", 74.29)
	DB.Add("GRMN", 144.37)
	DB.Add("CSL.AX", 288.18)
	DB.Add("QAN.AX", 4.55)
	DB.Add("WPL.AX", 22.45)

	// Run the fake DB updater every couple of seconds followed by the notifier
	go func() {
		for {
			time.Sleep(databaseUpdateFrequency)
			NF.Notify(DB, randomUpdater(DB))
		}
	}()

	// Defined the websocket HTTP handler for all inbound connections
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Connect from", r.RemoteAddr)
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "Bye")
		runClient(NF, DB, conn)
	})

	// Start up the websocket HTTP server
	log.Println(Name, "Listening on", args[0])
	log.Fatal(http.ListenAndServe(args[0], fn))
}

// runClient consumes message sent by the client. It is started as a separate go routine
// via the http.HandlerFunc setup.
//
// The read timeout is set to two times the ping interval which should give clients plenty
// of time to get a ping to us, even if they're busy. If no message arrives within that
// time, the client is considered unresponsive and discarded.
//
// Notification messages sent back to the client are handled by the Notifier.
//
// Strictly, a client need not send a TickerRequest message at all and simple Ping the
// server; or it can send multiple TickerRequest messages to replace eariler tickers of
// interest.
func runClient(NF *Notifier, DB *Database, conn *websocket.Conn) {
	var client *Client // Notifier correctly deals with a nil Client
	cid := ""
	for {
		ctx, cancel := context.WithTimeout(context.Background(), pingpongInterval*2)
		in, err := ReadMessage(ctx, conn)
		cancel() // Free up context resources
		if err != nil {
			log.Println(cid, err)
			break
		}
		switch msg := in.(type) {
		case *TickerRequest:
			NF.Delete(client) // Replace old entry (if any) with new
			client = &Client{id: msg.Id, tickers: msg.Tickers, conn: conn}
			cid = msg.Id
			NF.Add(client)

			// Add client tickers to the database if not already present. This
			// is mainly to make the demo easy to run.
			for ix, ticker := range msg.Tickers {
				DB.Add(ticker, float64(ix*ix))
			}

		case *Ping:
			log.Println(cid, "Ping")
			out := &Pong{Sequence: msg.Sequence,
				Seconds: msg.Seconds, Nanos: msg.Nanos}
			err := WriteMessage(context.Background(), conn, out)
			if err != nil {
				break
			}

		default:
			log.Printf(cid, "Unexpected Message Type %v", in)
			break
		}
	}

	NF.Delete(client)
}
