package main

import (
	"fmt"
	"os"

	_ "go-study/stdlib-http/advanced"
	"go-study/stdlib-http/client"
	_ "go-study/stdlib-http/server"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "client":
		runClientExamples()
	case "server":
		runServerExamples()
	case "advanced":
		runAdvancedExamples()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: go run main.go [command]")
	fmt.Println("\nCommands:")
	fmt.Println("  client    - Run HTTP client examples")
	fmt.Println("  server    - Run HTTP server examples")
	fmt.Println("  advanced  - Run advanced HTTP examples")
}

func runClientExamples() {
	fmt.Println("=== HTTP Client Examples ===")
	// In a real CLI app, you might use another argument to select one.
	// For simplicity, we'll just run them all here.
	fmt.Println("\n1. Basic GET request:")
	client.BasicGet()

	fmt.Println("\n2. GET with custom headers:")
	client.GetWithHeaders()

	fmt.Println("\n3. POST JSON data:")
	client.PostJSON()

	fmt.Println("\n4. Custom HTTP client with timeout:")
	client.CustomClient()
}

func runServerExamples() {
	fmt.Println("=== HTTP Server Examples ===")
	fmt.Println("Choose a server to run by uncommenting it in the code.")
	// To run an example, you would typically pass another argument to select it.
	// For now, you can uncomment one of the lines below to start a server.

	// server.BasicServer()
	// server.ServerWithMux()
	// server.ServerWithMiddleware()

	fmt.Println("\nTo run a server, edit 'runServerExamples' in main.go and uncomment a server function.")
}

func runAdvancedExamples() {
	fmt.Println("=== Advanced HTTP Examples ===")
	fmt.Println("Choose an example to run by uncommenting it in the code.")

	// advanced.ParseFormParams()
	// advanced.FileUpload()
	// advanced.FileServer()

	fmt.Println("\nTo run an advanced example, edit 'runAdvancedExamples' in main.go and uncomment a function.")
}
