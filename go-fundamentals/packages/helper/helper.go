package helper

import "strings"

// PublicFunction is an example of a function that can be exported.
// In Go, any function or variable that starts with a capital letter is exported (public).
func PublicFunction(text string) string {
	return strings.ToUpper(text)
}

// privateFunction is not visible outside the `helper` package
// because it starts with a lowercase letter.
func privateFunction() string {
	return "this is private"
}
