package demo

import "fmt"

type Speaker interface {
	Speak() string
}

type Dog struct{}

func (d Dog) Speak() string {
	return "Woof"
}

type cat struct{}

func (c cat) Speak() string {
	return "Meow"
}

func Say(s Speaker) {
	fmt.Println(s.Speak())
}

func main() {
	Say(Dog{})
	Say(cat{})
}
