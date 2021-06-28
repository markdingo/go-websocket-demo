package main

import (
	"bytes"
	"fmt"

	"encoding/json"
)

// The application "protocol" is embodied in the Encode/Decode functions which create the
// Protocol Data Units (aka messages) exchanged over a websocket. They rely on the
// standard encoding/json package for serialization/deserialization.
//
// The main issue with using json with websockets (in go at least) is that on the
// receiving side there is no easy way to determine the appropriate struct to give to
// json.Unmarshal(). To work around this the encoder prefixes the json payload with the
// message name, e.g.: `TicketRequest:{"Id": "SomeID"...`
//
// Conversely the decoder peels off the message name and uses it to find the matching
// struct constructor to create the struct needed by json.Unmarshal() to decode the
// remaining payload.

// messageConstructorFunc is a function which creates a new, empty message struct
// suitable for json.Unmarshal. It's indexed on the struct name by the registry.
type messageConstructorFunc func() Message

// registry associates message names with struct constructors
var registry = make(map[string]messageConstructorFunc)

// Register associates a message name with a message constructor. The registry is used by
// Decode to construct and deserialize the message into the matching struct.
func Register(name string, f messageConstructorFunc) {
	registry[name] = f
}

// Encode creates the message as a byte slice containing `messagename` + `:` + json-payload
func Encode(msg Message) []byte {
	n := make([]byte, 0, 1000) // Make a guess at the size needed - being wrong doesn't hurt
	n = []byte(msg.Name())
	n = append(n, Sep)
	b, _ := json.Marshal(msg)
	n = append(n, b...)

	return n
}

// Decode reverses Encode. An error is returned if there is no name prefix, the name
// refers to a non-existant message or if the json unmarshal fails.
func Decode(b []byte) (Message, error) {
	ix := bytes.IndexByte(b, Sep)
	if ix == -1 {
		return nil, fmt.Errorf("Message lacks a name prefix separated by '%b'", Sep)
	}

	name := string(b[:ix])
	payload := b[ix+1:]

	// Does this message have a constructor in the registry?
	constructor := registry[name]
	if constructor == nil {
		return nil, fmt.Errorf("Unknown Message Name '%s'", name)
	}

	msg := constructor()
	err := json.Unmarshal(payload, msg)

	return msg, err
}
