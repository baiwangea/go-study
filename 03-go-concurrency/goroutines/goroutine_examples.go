package goroutines

import (
	"fmt"
	"time"
)

func say(s string) {
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func DemonstrateGoroutines() {
	fmt.Println("--- Goroutine Demonstration ---")

	// Start a new goroutine by using the `go` keyword.
	// This will execute the `say` function concurrently with the `main` goroutine.
	go say("world")

	// The `main` goroutine continues its execution.
	say("hello")

	fmt.Println("\nNote: The execution order is not deterministic.")
	fmt.Println("The main goroutine finished, so the program might exit before the 'world' goroutine completes all its prints.")
	fmt.Println("This is why we need synchronization primitives like WaitGroups.")
}
