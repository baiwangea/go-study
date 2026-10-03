package channels

import (
	"fmt"
	"time"
)

// DemonstrateUnbufferedChannels shows how goroutines can communicate using an unbuffered channel.
func DemonstrateUnbufferedChannels() {
	fmt.Println("\n--- Unbuffered Channel Demonstration ---")

	// Create an unbuffered channel of strings.
	// Unbuffered channels require both a sender and a receiver to be ready at the same time.
	messages := make(chan string)

	go func() {
		fmt.Println("Goroutine: sending 'ping'...")
		// Send a value into the channel. This line will block until another goroutine receives from it.
		messages <- "ping"
		fmt.Println("Goroutine: 'ping' sent!")
	}()

	// Wait for a moment to show the goroutine is waiting.
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Main: waiting to receive...")

	// Receive a value from the channel. This line blocks until a value is sent.
	msg := <-messages
	fmt.Println("Main: received message:", msg)
}

// DemonstrateBufferedChannels shows how a buffered channel allows a certain number of sends without a ready receiver.
func DemonstrateBufferedChannels() {
	fmt.Println("\n--- Buffered Channel Demonstration ---")

	// Create a buffered channel with a capacity of 2.
	// The sender can send up to 2 values before it blocks.
	messages := make(chan string, 2)

	// Send two values to the channel without blocking.
	messages <- "buffered"
	messages <- "channel"
	fmt.Println("Sent two messages to the buffered channel.")

	// Now, receiving the values.
	fmt.Println("Received:", <-messages)
	fmt.Println("Received:", <-messages)
}

// DemonstrateChannelDirections shows how to specify if a channel is meant for sending or receiving.
func DemonstrateChannelDirections() {
	fmt.Println("\n--- Channel Directions (Types) Demonstration ---")

	pings := make(chan string, 1)
	pongs := make(chan string, 1)

	ping(pings, "hello world")
	pong(pings, pongs)

	fmt.Println(<-pongs)
}

// ping function only accepts a channel for sending.
func ping(pings chan<- string, msg string) {
	fmt.Println("Sending a message to the ping channel...")
	pings <- msg
}

// pong function accepts one channel for receiving (pings) and another for sending (pongs).
func pong(pings <-chan string, pongs chan<- string) {
	msg := <-pings
	fmt.Println("Received a message from the ping channel.")
	fmt.Println("Sending the message to the pong channel...")
	pongs <- msg
}
