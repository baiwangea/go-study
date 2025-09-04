package mutexes

import (
	"fmt"
	"sync"
	"time"
)

// SafeCounter is a counter that is safe to use concurrently.
// It embeds a sync.Mutex to protect its internal state.
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

// Inc increments the counter in a thread-safe way.
func (c *SafeCounter) Inc() {
	c.mu.Lock()         // Lock the mutex before accessing the shared resource.
	defer c.mu.Unlock() // Unlock the mutex when the function returns.
	c.count++
}

// Value returns the current value of the counter in a thread-safe way.
func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func DemonstrateMutex() {
	fmt.Println("\n--- Mutex Demonstration ---")

	c := SafeCounter{}
	// Start 1000 goroutines that all increment the counter.
	for i := 0; i < 1000; i++ {
		go c.Inc()
	}

	// Wait for a moment to allow goroutines to run.
	time.Sleep(time.Second)

	// Without the mutex, the final count would likely be less than 1000
	// due to race conditions where multiple goroutines read and write the same value.
	fmt.Println("Final counter value:", c.Value())
}
