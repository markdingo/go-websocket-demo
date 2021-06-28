package main

const (
	Sep = ':'
)

// Message is the interface all message structs must meet so that they can be
// Encoded/Decoded.
type Message interface {
	Name() string
}

// init registers the constructors for all known message types
func init() {
	Register("TickerRequest", func() Message { return new(TickerRequest) })
	Register("TickerChange", func() Message { return new(TickerChange) })
	Register("Ping", func() Message { return new(Ping) })
	Register("Pong", func() Message { return new(Pong) })
}

// TickerRequest is sent by the client identify which symbols tickers it cares about.
type TickerRequest struct {
	Id      string   // A printable identifier created by the client to help self-identify
	Tickers []string // List of tickers the client wants to be notified about
}

func (t *TickerRequest) Name() string {
	return "TickerRequest"
}

type SymbolChange struct {
	Symbol     string
	UpdateTime int64   // time.Unix() so it survives json
	Price      float64 // Cents
}

// TickerChange is sent by the server to indicate which tickers have changed
type TickerChange struct {
	Changes []SymbolChange
}

func (t *TickerChange) Name() string {
	return "TickerChange"
}

// Ping is sent by the client to indicate it is still responsive
type Ping struct {
	Sequence int
	Seconds  int64 // time.Unix() so it survives json
	Nanos    int64 // time.UnixNano()

}

func (t *Ping) Name() string {
	return "Ping"
}

// Pong is the server response to a Ping message. It indicates that the server is still
// responsive.
type Pong struct {
	Sequence int
	Seconds  int64
	Nanos    int64
}

func (t *Pong) Name() string {
	return "Pong"
}
