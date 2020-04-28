package Pop

import (
	"fmt"

	_ "github.com/reactivego/generics/example/stack/generic"
)

func Example() {
	s := StringStack{"Hello, World!"}

	if value, ok := s.Pop(); ok {
		fmt.Println(value)
	} else {
		fmt.Println("empty")
	}

	if value, ok := s.Pop(); ok {
		fmt.Println(value)
	} else {
		fmt.Println("empty")
	}

	// Output:
	// Hello, World!
	// empty
}
