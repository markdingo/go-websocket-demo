package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

const (
	Name = "client"
)

func main() {
	myId := fmt.Sprintf("%s:%d ", Name, os.Getpid())
	log.SetPrefix(myId)
	log.SetFlags(log.Ltime)
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Usage:", Name, "serverURL ticker1 [ticker2...tickern]")
		return
	}
	serverURL := args[0]
	args = args[1:] // Shift

	// Iterate on server timeouts and errors, but give it a breather if it does fail
	// so that we're not hammering against a dead or recovering server.
	for {
		err := askServer(myId, serverURL, args)
		if err != nil {
			log.Println("Warning:", err.Error())
		}
		time.Sleep(serverRetryConnectDelay) // Breather
	}
}

// askServer connects to the websocket server and sends a ticker update request
// message. Return on error or closed socket and let caller retry.
func askServer(myId, serverURL string, tickers []string) (err error) {
	log.Println("Asking", serverURL, "to watch:", strings.Join(tickers, ","))

	dialCtx, cancel := context.WithTimeout(context.Background(), serverDialTimeout)
	defer cancel()

	conn, _, err := websocket.Dial(dialCtx, serverURL, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Connection is established. Create a parent connection context so that all
	// interested parties know when the connection dies.
	connCtx, connCancel := context.WithCancel(context.Background())
	defer connCancel()

	// Send the server our tickers of interest.

	out := &TickerRequest{Id: myId, Tickers: tickers}
	err = WriteMessage(connCtx, conn, out)
	if err != nil {
		return
	}

	go pingPong(connCtx, cancel, conn) // Start application level ping/pong exchange

	// Loop forever reading server messages. A timeout is fatal as it means that the
	// ping/pong exchange has failed.

	for {
		ctx, cancel := context.WithTimeout(connCtx, pingpongInterval*2)
		in, err := ReadMessage(ctx, conn)
		cancel()
		if err != nil {
			return err
		}

		switch msg := in.(type) { // Dispatch on the messages we handle
		case *TickerChange:
			s := ""
			for _, change := range msg.Changes {
				s += fmt.Sprintf(" %s at $%0.2f", change.Symbol, change.Price)
			}
			log.Printf("Change notification(s): (%d)%s", len(msg.Changes), s)
		case *Pong:
			sentAt := time.Unix(msg.Seconds, msg.Nanos)
			latency := time.Now().Sub(sentAt)
			log.Println("Pong", msg.Sequence, "latency", latency)
		default:
			return fmt.Errorf("Unexpected Message Type %v", in)
		}
	}
}

// pingPong is started as a separate go routine which exchanges application-level Ping
// messages with the server until the context tells it to disappear. Any i/o error result
// in cancelling the parent context which notifies all interested parties - in this case
// really just the reader loop. The big-picture is that the websocket will fail if the
// ping exchange fails.
//
// Why is an application-level ping used rather than the intrinsic websocket ping?  Mainly
// because the websocket package doesn't expose the Pong message making it impossible for
// a server to determine that a client is unresponsive. Unfortunate but true.
//
// Mind you, there is nothing magic about the websocket ping messages. They are just a
// different message type, so providing our own version in the application layer is not
// any less efficient and supporting it in the server is a mere matter of a few lines of
// code, so no big deal.
func pingPong(parentCtx context.Context, parentCancel context.CancelFunc, conn *websocket.Conn) {
	defer parentCancel()
	defer log.Println("pingPong Exit")

	for seq := 0; ; seq++ {
		select {
		case <-parentCtx.Done(): // Did the reader loop cancel the context?
			return

		case now := <-time.After(pingpongInterval): // Or is it time for a Ping?
			msg := &Ping{Sequence: seq,
				Seconds: now.Unix(), Nanos: int64(now.Nanosecond())}
			log.Println("Ping", seq)
			err := WriteMessage(parentCtx, conn, msg)
			if err != nil {
				log.Println("pingPong i/o error", err)
				return
			}
		}
	}
}
