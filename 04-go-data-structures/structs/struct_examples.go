package structs

import "fmt"

// 1. Defining a struct
// A struct is a collection of fields.
// It's used to group related data together to form a more complex data type.
type Person struct {
	Name string
	Age  int
}

// 4. Methods
// A method is a function with a special receiver argument.
// Here, `(p Person)` is the receiver. It attaches the function `Introduce` to the `Person` type.
func (p Person) Introduce() {
	fmt.Printf("Hi, my name is %s and I am %d years old.\n", p.Name, p.Age)
}

// This method has a pointer receiver. It can modify the value to which the receiver points.
func (p *Person) HaveBirthday() {
	p.Age++
}

func DemonstrateStructs() {
	fmt.Println("\n--- Struct Demonstrations ---")

	// 2. Creating instances (instantiation)
	fmt.Println("\n2. Creating Instances:")
	// You can create an instance by specifying the values for the fields.
	p1 := Person{Name: "Alice", Age: 28}
	fmt.Printf("p1: %+v\n", p1) // %+v prints both field names and values

	// You can also use the `new` keyword, which returns a pointer to the struct.
	p2 := new(Person)
	p2.Name = "Bob"
	p2.Age = 35
	fmt.Printf("p2 (a pointer): %+v\n", *p2)

	// 3. Accessing fields
	fmt.Println("\n3. Accessing Fields:")
	fmt.Printf("The name of p1 is %s.\n", p1.Name)

	// Using the method attached to the Person type.
	fmt.Println("\n4. Using Methods:")
	p1.Introduce()
	p2.Introduce()

	// Using the pointer receiver method.
	fmt.Println("\n5. Using Pointer Receiver Methods:")
	fmt.Printf("p1's age before birthday: %d\n", p1.Age)
	p1.HaveBirthday()
	fmt.Printf("p1's age after birthday: %d\n", p1.Age)
}
