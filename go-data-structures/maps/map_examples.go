package maps

import "fmt"

func DemonstrateMaps() {
	fmt.Println("\n--- Map Demonstrations ---")

	// 1. Declaration and Initialization
	// A map stores key-value pairs. The zero value of a map is nil.
	// A nil map has no keys, nor can keys be added.
	fmt.Println("\n1. Declaration & Initialization:")
	m := make(map[string]int) // Using the make function

	// 2. Setting and Getting values
	fmt.Println("\n2. Setting & Getting:")
	m["k1"] = 7
	m["k2"] = 13
	fmt.Println("map:", m)

	v1 := m["k1"]
	fmt.Println("v1:", v1)

	// 3. The "comma ok" idiom
	// When getting a value from a map, you can get a second boolean value
	// which indicates if the key was actually present in the map.
	fmt.Println("\n3. The 'comma ok' idiom:")
	_, prs := m["k2"] // prs (present)
	fmt.Println("Is 'k2' present?:", prs)

	v3, prs := m["k3"] // k3 is not in the map
	fmt.Println("Is 'k3' present?:", prs, "Value:", v3, "(zero value for int)")

	// 4. Deleting a key-value pair
	fmt.Println("\n4. Deleting:")
	delete(m, "k2")
	fmt.Println("map after deleting 'k2':", m)

	// 5. Iterating
	// Use a for...range loop to iterate over a map.
	// Note: The iteration order of maps is not guaranteed!
	fmt.Println("\n5. Iterating with for...range:")
	data := map[string]string{"a": "apple", "b": "banana"}
	for k, v := range data {
		fmt.Printf("  %s -> %s\n", k, v)
	}
}
