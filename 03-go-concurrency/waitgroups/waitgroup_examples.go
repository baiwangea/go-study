package waitgroups

import (
	"fmt"
	"sync"
	"time"
)

// worker is a function that simulates some work.
func worker(id int, wg *sync.WaitGroup) {
	// Decrement the WaitGroup counter when the goroutine completes.
	defer wg.Done()

	fmt.Printf("Worker %d starting\n", id)

	// Simulate some work.
	time.Sleep(time.Second)
	fmt.Printf("Worker %d done\n", id)
}

func DemonstrateWaitGroup() {
	fmt.Println("\n--- WaitGroup Demonstration ---")

	// Create a new WaitGroup.
	var wg sync.WaitGroup

	// Launch several goroutines and increment the WaitGroup counter for each.
	for i := 1; i <= 3; i++ {
		wg.Add(1) // Increment the counter.
		go worker(i, &wg)
	}

	// Block until the WaitGroup counter goes back to 0.
	// This happens when all goroutines have called wg.Done().
	wg.Wait()

	fmt.Println("All workers have finished.")
}
