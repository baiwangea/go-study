package sorting

import "fmt"

// BubbleSort sorts a slice of integers using the bubble sort algorithm.
// It repeatedly steps through the list, compares adjacent elements and swaps them if they are in the wrong order.
// The pass through the list is repeated until the list is sorted.
func BubbleSort(arr []int) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		// Last i elements are already in place
		for j := 0; j < n-i-1; j++ {
			// Swap if the element found is greater than the next element
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

func DemonstrateBubbleSort() {
	fmt.Println("\n--- Bubble Sort Demonstration ---")
	arr := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("Unsorted array:", arr)
	BubbleSort(arr)
	fmt.Println("Sorted array:  ", arr)
}
