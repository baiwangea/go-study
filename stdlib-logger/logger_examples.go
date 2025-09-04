package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// BasicLogging demonstrates the standard logging functions.
func BasicLogging() {
	fmt.Println("--- Basic Logging ---")
	log.Println("This is a standard log message.")
	log.Printf("You can also format messages, like this number: %d", 123)

	// By default, log prints to stderr.
}

// LogToFile demonstrates how to redirect log output to a file.
func LogToFile() {
	fmt.Println("\n--- Logging to a File ---")

	// Open a file for writing. Create it if it doesn't exist.
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	// Set the output of the standard logger to the file.
	log.SetOutput(file)

	log.Println("This message will be written to app.log.")
	log.Println("And this one too.")

	// Reset the output to the default (stderr) for other examples.
	log.SetOutput(os.Stderr)
	fmt.Println("Check the 'app.log' file for the output.")
}

// CustomizeLogger demonstrates how to change the prefix and flags of the logger.
func CustomizeLogger() {
	fmt.Println("\n--- Customizing the Logger ---")

	// Flags define what information to include in the log entry.
	// Ldate | Ltime = Date and Time
	// Lshortfile = File name and line number
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("[MyApp] ")

	log.Println("This is a log message with a custom prefix and flags.")

	// Reset to default for other examples
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")
}

// CreateNewLogger demonstrates creating a new, independent logger instance.
func CreateNewLogger() {
	fmt.Println("\n--- Creating a New Logger Instance ---")

	var buf bytes.Buffer

	// Create a new logger that writes to the buffer.
	// It has a custom prefix and includes the long file name and line number.
	logger := log.New(&buf, "[Info] ", log.Llongfile)

	logger.Println("This is a message from our custom logger.")
	logger.Println("It writes to an in-memory buffer.")

	fmt.Println("The custom logger's buffer contains:")
	fmt.Print(buf.String())
}

// FatalAndPanic demonstrate logging functions that stop program execution.
func FatalAndPanic() {
	fmt.Println("\n--- Fatal and Panic (Demonstration) ---")
	fmt.Println("The following examples will stop the program if uncommented.")

	// log.Fatal is equivalent to log.Print() followed by os.Exit(1).
	// fmt.Println("About to call log.Fatal...")
	// log.Fatal("This is a fatal error. The program will now exit.")
	// fmt.Println("This line will not be reached.")

	// log.Panic is equivalent to log.Print() followed by a panic().
	// It can be recovered from using a defer/recover mechanism.
	// fmt.Println("About to call log.Panic...")
	// log.Panic("This is a panic. The program will stop unless recovered.")
	// fmt.Println("This line will also not be reached.")
}
