package main

import "fmt"

// 1. Pass by Value
// This function receives a COPY of the integer value.
// Any changes made to `val` inside this function will not affect the original variable.
func zeroVal(ival int) {
	ival = 0
}

// 2. Pass by Pointer
// This function receives a POINTER to an integer.
// By dereferencing the pointer, we can modify the original variable's value.
func zeroPtr(iptr *int) {
	*iptr = 0 // The * operator dereferences the pointer, giving us access to the value it points to.
}

func main() {
	fmt.Println("====== Go Pointers Demonstration ======")

	i := 1
	fmt.Println("Initial value of i:", i)

	// --- Pass by Value ---
	zeroVal(i)
	fmt.Println("Value of i after zeroVal (pass by value):", i, "(un-changed)")

	// --- Pass by Pointer ---
	// The &i syntax gives us the memory address of i, i.e., a pointer to i.
	zeroPtr(&i)
	fmt.Println("Value of i after zeroPtr (pass by pointer):", i, "(changed)")

	// --- Pointer Basics ---
	// The type *int is a pointer to an int.
	// The zero value of a pointer is nil.
	var p *int
	fmt.Println("\nZero value of a pointer:", p)

	// You can get the memory address of a variable using the & operator.
	fmt.Println("Memory address of i:", &i)

	// You can see that the pointer `&i` now holds the value 0, because we changed it via zeroPtr.
	fmt.Println("Value at that memory address (dereferenced):", *&i)

	fmt.Println("\n====== Pointers Demonstration Complete ======")
}
