package slices

import "fmt"

func DemonstrateSlices() {
	fmt.Println("--- Slice Demonstrations ---")

	// 1. Declaration and Initialization
	// A slice is a dynamically-sized, flexible view into the elements of an array.
	fmt.Println("\n1. Declaration & Initialization:")
	s := []string{"a", "b", "c"} // A slice literal
	fmt.Printf("s = %v, len=%d, cap=%d\n", s, len(s), cap(s))

	// 2. Appending
	// The append function is used to add elements to the end of a slice.
	fmt.Println("\n2. Appending:")
	s = append(s, "d")
	s = append(s, "e", "f")
	fmt.Printf("After append: s = %v, len=%d, cap=%d\n", s, len(s), cap(s))
	// Notice how the capacity might grow larger than the length.

	// 3. Slicing a slice (creating a sub-slice)
	// You can create a new slice that points to a section of an existing slice.
	// Format: slice[low:high]
	fmt.Println("\n3. Slicing a slice:")
	l := s[2:5] // Elements from index 2 up to (but not including) 5
	fmt.Printf("s[2:5] -> l = %v\n", l)

	l = s[:5] // Elements from the beginning up to 5
	fmt.Printf("s[:5]  -> l = %v\n", l)

	l = s[2:] // Elements from index 2 to the end
	fmt.Printf("s[2:]  -> l = %v\n", l)

	// 4. Iterating
	// The `for...range` loop is the idiomatic way to iterate over a slice.
	fmt.Println("\n4. Iterating with for...range:")
	for i, v := range s {
		fmt.Printf("  Index %d: %s\n", i, v)
	}

	// 5. Copying
	// To create a truly independent slice, you must use the `copy` function.
	fmt.Println("\n5. Copying a slice:")
	c := make([]string, len(s))
	numCopied := copy(c, s)
	fmt.Printf("Copied %d elements into a new slice c = %v\n", numCopied, c)

	// Modifying the copy does not affect the original.
	c[0] = "MODIFIED"
	fmt.Printf("Original s = %v\n", s)
	fmt.Printf("Modified c = %v\n", c)
}
