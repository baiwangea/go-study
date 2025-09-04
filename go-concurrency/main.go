package main

import (
	"fmt"
	"go-study/go-concurrency/channels"
	"go-study/go-concurrency/goroutines"
	"go-study/go-concurrency/mutexes"
	"go-study/go-concurrency/selects"
	"go-study/go-concurrency/waitgroups"
)

func main() {
	fmt.Println("====== Go Concurrency Patterns ======")

	// Basic goroutine usage
	goroutines.DemonstrateGoroutines()

	// Using WaitGroup for synchronization
	waitgroups.DemonstrateWaitGroup()

	// Using Mutex for safe concurrent access
	mutexes.DemonstrateMutex()

	// Channel communication
	channels.DemonstrateUnbufferedChannels()
	channels.DemonstrateBufferedChannels()
	channels.DemonstrateChannelDirections()

	// Using select for multi-channel operations
	selects.DemonstrateSelect()
	selects.DemonstrateSelectWithTimeout()

	fmt.Println("\n====== Concurrency Demonstrations Complete ======")
}
