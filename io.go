package main

import (
	"context"
	"fmt"
	"io"

	"nhooyr.io/websocket"
)

// WriteMessage serializes a single Message and writes it to the websocket
// connection. WriteMessage sets a context Timeout to ensure the websocket package doesn't
// stall forever.
func WriteMessage(parentCtx context.Context, conn *websocket.Conn, msg Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	n := Encode(msg)
	return conn.Write(ctx, websocket.MessageText, n)
}

// ReadMessage waits for a Message from the server. On receipt the message is decoded
// into the corresponding struct. The caller must ensure that the provided context has a
// timeout if they want the websocket.Conn to handle timeouts directly.
func ReadMessage(ctx context.Context, conn *websocket.Conn) (Message, error) {
	mt, r, err := conn.Reader(ctx)
	if err != nil {
		return nil, err
	}

	if mt != websocket.MessageText {
		return nil, fmt.Errorf("Unexpected message type %s", mt.String())
	}
	b, err := io.ReadAll(r)
	if err == nil {
		msg, err := Decode(b)
		return msg, err
	}

	return nil, err
}
