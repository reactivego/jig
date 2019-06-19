package main

import _ "github.com/reactivego/rx/generic"

func main() {
	FromStrings("You!", "Gophers!", "World!").
		MapString(func(x string) string {
			return "Hello, " + x
		}).
		Println()

	// Ouptput:
	// Hello, You!
	// Hello, Gophers!
	// Hello, World!
}