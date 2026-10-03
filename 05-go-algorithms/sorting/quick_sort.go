package sorting

import "fmt"

// QuickSort sorts a slice of integers using the quick sort algorithm.
// It picks an element as a pivot and partitions the given array around the picked pivot.
func QuickSort(arr []int) {
	quickSortRecursive(arr, 0, len(arr)-1)
}

func quickSortRecursive(arr []int, low, high int) {
	if low < high {
		// pi is partitioning index, arr[pi] is now at right place
		pi := partition(arr, low, high)

		// Separately sort elements before partition and after partition
		quickSortRecursive(arr, low, pi-1)
		quickSortRecursive(arr, pi+1, high)
	}
}

// This function takes last element as pivot, places the pivot element at its correct
// position in sorted array, and places all smaller (smaller than pivot) to left of
// pivot and all greater elements to right of pivot
func partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1 // Index of smaller element

	for j := low; j < high; j++ {
		// If current element is smaller than or equal to pivot
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i] // swap
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1] // swap pivot to its correct position
	return i + 1
}

func DemonstrateQuickSort() {
	fmt.Println("\n--- Quick Sort Demonstration ---")
	arr := []int{10, 7, 8, 9, 1, 5}
	fmt.Println("Unsorted array:", arr)
	QuickSort(arr)
	fmt.Println("Sorted array:  ", arr)
}
