package interfaces

import (
	"fmt"
	"math"
)

// 1. Defining an interface
// An interface type is defined as a set of method signatures.
// Here, any type that has a method `Area() float64` implicitly implements the `Shape` interface.
type Shape interface {
	Area() float64
}

// 2. Implementing the interface with concrete types

// `Rect` struct
type Rect struct {
	Width, Height float64
}

// `Circle` struct
type Circle struct {
	Radius float64
}

// `Rect` implements the `Shape` interface by defining an `Area()` method.
func (r Rect) Area() float64 {
	return r.Width * r.Height
}

// `Circle` also implements the `Shape` interface.
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// 3. Using the interface
// This function takes any `Shape` as an argument. It doesn't need to know
// whether it's a Rect, a Circle, or any other type, as long as it has an `Area()` method.
func measure(s Shape) {
	fmt.Printf("Shape: %T, Area: %0.2f\n", s, s.Area())
}

func DemonstrateInterfaces() {
	fmt.Println("\n--- Interface Demonstrations ---")

	r := Rect{Width: 10, Height: 5}
	c := Circle{Radius: 5}

	// We can call `measure` with both `r` and `c` because they both satisfy the `Shape` interface.
	measure(r)
	measure(c)
}
