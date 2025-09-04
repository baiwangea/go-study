package selects

import (
	"fmt"
	"time"
)

func DemonstrateSelect() {
	fmt.Println("\n--- Select Demonstration ---")

	// Create two channels.
	c1 := make(chan string)
	c2 := make(chan string)

	// Start a goroutine that sends a value to c1 after 2 seconds.
	go func() {
		time.Sleep(2 * time.Second)
		c1 <- "one"
	}()

	// Start another goroutine that sends a value to c2 after 1 second.
	go func() {
		time.Sleep(1 * time.Second)
		c2 <- "two"
	}()

	// Use a for loop to wait for both messages.
	for i := 0; i < 2; i++ {
		fmt.Println("Waiting for a message...")
		// The select statement blocks until one of its cases can run.
		// If multiple cases are ready, it chooses one at random.
		select {
		case msg1 := <-c1:
			fmt.Println("Received from c1:", msg1)
		case msg2 := <-c2:
			fmt.Println("Received from c2:", msg2)
		}
	}
}

func DemonstrateSelectWithTimeout() {
	fmt.Println("\n--- Select with Timeout Demonstration ---")

	c1 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		c1 <- "result 1"
	}()

	// Here we use select with a timeout.
	// `time.After` creates a channel that will receive a value after the specified duration.
	select {
	case res := <-c1:
		fmt.Println(res)
	case <-time.After(1 * time.Second):
		fmt.Println("Timeout: Did not receive a result within 1 second.")
	}
}
