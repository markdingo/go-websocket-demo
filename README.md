# go-websocket-demo

### Introduction

Demo go code showing a simple websocket client connecting to a simple websocket server
asking for a feed of stock ticker changes using the recommended
[nhooyr.io/websocket](https://nhooyr.io/websocket) package.

The server maintains the websocket connection so long as the client periodically sends
Ping messages indicating responsiveness. Similarly the client maintains the connection so
long as the server responds with a Pong message indicating the server is also
responsive. In other words: clients know when servers become unresponsive and vice versa
and thus both sides can release resources and retry as appropriate.

Note that the Ping/Pong messages are application-level messages rather than the intrinsic
message types available within most websocket packages. The reason for this, in part, is
that the recommended websocket package only exposed Ping, but not Pong (rather odd - but
an issue has been raised, so maybe that will change one day). Point being that that makes
a server incapable of determining whether a client is still responsive or not. Sure, the
client can tell if the server is unresponsive, but that's only half the story. In our
scenario it's important for the server to know when a client is unresponsive so it doesn't
continue to consume resources unnecessarily.

The second reason for using application-level Pings is to demonstrate message demuxing
with json as the serialization format. This is not an overly complex problem but there are
nuances due to the nature of go and the requirements of the `encoding/json` package. In
short, if you just exchange pure json you need to construct the message struct before you
know the message type... Thus the exchanged messages include struct identifiers. More
details can be found in [protocol.go](https://github.com/markdingo/go-websocket-demo/blob/main/protocol.go).

### Purpose

The main goal of this package is to demonstrate a performant and responsive websocket
server implemented in go. At the time of writing the whole system including client, server
and support modules adds up to less than 400 executable lines of code yet the server
efficiently handles multiple clients, includes an extensible framework which allows the
easy addition of new message types and actively deals with unresponsive clients.

### Caveat Emptor

This code is demo quality only: there is no client authentication by the server; there are
no units tests and this code has never been used in earnest.

### How to use

Assuming a Unix platform:

1. `git clone git@github.com:markdingo/go-websocket-demo.git` or if I've made the repo
public possibly `go get -u github.com/markdingo/go-websocket-demo`

1. Build the client and server with `'make'`

1. Run `'./server localhost:8080'` in one terminal

1. Run `'./client ws://localhost:8080 AAPL SPY'` in another terminal

1. Repeat step 3 as many times as you wish in new terminal windows

1. If you wait for over a minute you'll see that client exchange Ping messages with the server
and print the latency of that message exchange

1. Simulate a stalled server by typing `^Z` in the server terminal and watch the clients
eventually determine that the server is unresponsive and attempt reconnections. The Ping
timeout is set to one minute so clients will notice an unresponsive server within two
minutes. Modify `config.go` if you want to change this or other timeout values.

1. Revivify the server by typing `fg` in the server terminal and watch the clients
automatically re-connect and continue receiving stock symbol notifications

1. Do the same thing with the clients: `^Z` to stall them, then watch the server
ultimately discard unresponsive clients. Then `fg` the client and watch it reconnect
and resume getting ticker notifications.

### Runs on?

This demo is know to work on Linux, FreeBSD and macOS using go1.16 or later.

**--**

### Copyright and License

go-websocket-demo is Copyright :copyright: 2021 Mark Delany. This software is licensed
under the BSD 2-Clause "Simplified" License.
