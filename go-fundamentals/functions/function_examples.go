package functions

import (
	"fmt"
	"strings"
)

// 1. Basic function with parameters and a return value.
func plus(a int, b int) int {
	return a + b
}

// 2. Multiple return values.
// Go functions can return multiple results.
func vals() (int, int) {
	return 3, 7
}

// 3. Variadic functions.
// A function that can be called with any number of trailing arguments.
func sum(nums ...int) (int, int) {
	fmt.Print(nums, " -> ")
	total := 0
	for _, num := range nums {
		total += num
	}
	return len(nums), total
}

// 4. Closures (Anonymous Functions).
// Go supports anonymous functions, which can form closures.
// The function `intSeq` returns another function, which we define anonymously.
// The returned function `closes over` the variable `i` to form a closure.
func intSeq() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}

// 5. Recursion.
func fact(n int) int {
	if n == 0 {
		return 1
	}
	return n * fact(n-1)
}

func DemonstrateFunctions() {
	fmt.Println("--- Function Demonstrations ---")

	// 1. Basic function call
	fmt.Println("\n1. Basic Function:")
	res := plus(1, 2)
	fmt.Println("1 + 2 =", res)

	// 2. Multiple return values
	fmt.Println("\n2. Multiple Return Values:")
	a, b := vals()
	fmt.Println("Returned values:", a, b)

	// 3. Variadic functions
	fmt.Println("\n3. Variadic Functions:")
	count, total := sum(1, 2)
	fmt.Printf("Summed %d numbers, total: %d\n", count, total)
	count, total = sum(1, 2, 3, 4)
	fmt.Printf("Summed %d numbers, total: %d\n", count, total)

	// 4. Closures
	fmt.Println("\n4. Closures (Anonymous Functions):")
	nextInt := intSeq()
	fmt.Println(nextInt()) // 1
	fmt.Println(nextInt()) // 2
	fmt.Println(nextInt()) // 3

	// 5. Recursion
	fmt.Println("\n5. Recursion:")
	fmt.Println("Factorial of 7 is", fact(7))
}
