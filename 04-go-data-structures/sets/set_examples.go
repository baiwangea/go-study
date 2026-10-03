package sets

import "fmt"

// 1. Defining a Set type
// We can define a Set as a map where the values are empty structs.
// `struct{}` is a zero-memory value, making it more efficient than `bool`.
type Set map[string]struct{}

// NewSet creates and returns a new Set.
func NewSet() Set {
	return make(Set)
}

// Add inserts an element into the set.
func (s Set) Add(element string) {
	s[element] = struct{}{}
}

// Remove deletes an element from the set.
func (s Set) Remove(element string) {
	delete(s, element)
}

// Contains checks if an element is in the set.
func (s Set) Contains(element string) bool {
	_, exists := s[element]
	return exists
}

// List returns a slice of all elements in the set.
func (s Set) List() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

func DemonstrateSets() {
	fmt.Println("\n--- Set Implementation Demonstrations ---")

	// 2. Using the Set
	fmt.Println("\n2. Using the Set:")
	mySet := NewSet()

	mySet.Add("apple")
	mySet.Add("banana")
	mySet.Add("apple") // Adding a duplicate element has no effect

	fmt.Println("Set contents:", mySet.List())

	fmt.Println("Does the set contain 'banana'?", mySet.Contains("banana"))
	fmt.Println("Does the set contain 'orange'?", mySet.Contains("orange"))

	mySet.Remove("apple")
	fmt.Println("Set contents after removing 'apple':", mySet.List())
}
