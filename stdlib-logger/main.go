package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Println("This program doesn't take arguments. It runs a sequence of demonstrations.")
		return
	}

	// Run all logging demonstrations.
	BasicLogging()
	LogToFile()
	CustomizeLogger()
	CreateNewLogger()
	FatalAndPanic()

	fmt.Println("\nLogging demonstrations complete.")
}
