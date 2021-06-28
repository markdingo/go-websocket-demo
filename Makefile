ALL=client server

all: $(ALL)

COMMONGO=io.go config.go messages.go protocol.go
COMMONDEP=$(COMMONGO) Makefile

client: client.go $(COMMONDEP)
	go build -o client client.go $(COMMONGO)

server: server.go $(COMMONDEP) database.go notifier.go
	go build -o server server.go $(COMMONGO) database.go notifier.go

# Compile with race-detector to ensure no concurrency issues
server-race: server.go $(COMMONDEP) database.go notifier.go
	go build -race -o server server.go $(COMMONGO) database.go notifier.go

fmt:
	gofmt -s -w `ls *.go`

clean:
	rm -f $(ALL)
