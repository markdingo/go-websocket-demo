package main

import (
	"time"
)

// In a real server you might want to determine these from an external config file or
// possibly command-line arguments. For a demo hard-coded is plenty good enough.

const (
	databaseUpdateFrequency = 2 * time.Second  // DB symbols random update period
	pingpongInterval        = 1 * time.Minute  // Client sends a ping
	serverDialTimeout       = 10 * time.Second // Client waits for a server connection
	serverRetryConnectDelay = 5 * time.Second  // Breather delay to avoid hammering servers
	writeTimeout            = 5 * time.Second  // websocket write must complete by this time
)
