package main

import (
	"fmt"
	"go-study/go-fundamentals/functions"
	"go-study/go-fundamentals/interfaces"
	"go-study/go-fundamentals/packages"
)

func main() {
	fmt.Println("====== Go Fundamentals ======")

	// Function examples
	functions.DemonstrateFunctions()

	// Package examples
	packages.DemonstratePackages()

	// Interface examples
	interfaces.DemonstrateInterfaces()

	fmt.Println("\n====== Fundamentals Demonstrations Complete ======")
}
