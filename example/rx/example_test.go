package rx

import "github.com/reactivego/rx"

/*
Use the generics from package "github.com/reactivego/rx" and generate
function FromStrings and method MapString and Println by running the
jig command.
*/
func Example_generic() {
	FromStrings("You!", "Gophers!", "World!").
		MapString(func(x string) string {
			return "Hello, " + x
		}).
		Println()

	// Output:
	// Hello, You!
	// Hello, Gophers!
	// Hello, World!
}


/*
Use the implementations from package "github.com/reactivego/rx" directly.
The rx package at the root of the library contains all generics expanded
for the interface{} type.
*/
func Example_heterogeneous() {
	type any = interface{}

	rx.From("You!", "Gophers!", "World!").
		Map(func(x any) any {
			return "Hello, " + x.(string)
		}).
		Println()

	// Output:
	// Hello, You!
	// Hello, Gophers!
	// Hello, World!
}
