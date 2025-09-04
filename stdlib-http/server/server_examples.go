package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// helloHandler is a simple handler that writes a welcome message.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, from the basic Go server!")
}

// BasicServer starts a simple HTTP server on port 8080.
func BasicServer() {
	// Register the handler function for the root path.
	http.HandleFunc("/", helloHandler)

	fmt.Println("Starting a basic server on http://localhost:8080")
	// Start the server. If it fails, log the error.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// ServerWithMux demonstrates how to use a ServeMux to handle different routes.
func ServerWithMux() {
	// Create a new ServeMux (router).
	mux := http.NewServeMux()

	// Register handlers for different routes.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		// In a real app, you would fetch and return user data.
		fmt.Fprintf(w, "Here are the users.")
	})

	fmt.Println("Starting a server with a mux on http://localhost:8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// loggingMiddleware is a middleware that logs information about each request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)

		log.Printf("Completed in %v", time.Since(start))
	})
}

// protectedHandler is a handler that simulates a protected resource.
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is a protected resource.")
}

// ServerWithMiddleware demonstrates how to chain middleware with a handler.
func ServerWithMiddleware() {
	// Create a new ServeMux.
	mux := http.NewServeMux()

	// The handler for the protected route.
	protected := http.HandlerFunc(protectedHandler)

	// Chain the logging middleware with the protected handler.
	mux.Handle("/protected", loggingMiddleware(protected))

	fmt.Println("Starting a server with middleware on http://localhost:8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
