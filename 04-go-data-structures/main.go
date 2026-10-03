package main

import (
	"go-study/go-data-structures/maps"
	"go-study/go-data-structures/sets"
	"go-study/go-data-structures/slices"
	"go-study/go-data-structures/structs"
)

func main() {
	// Run all data structure demonstrations in sequence.
	slices.DemonstrateSlices()
	maps.DemonstrateMaps()
	structs.DemonstrateStructs()
	sets.DemonstrateSets()
}
