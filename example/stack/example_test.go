// Package stack offers a heterogeneous Stack type.
package stack

import (

	_ "github.com/reactivego/jig/example/stack/generic"
)

func Example() {
	var s StringStack
	s.Push("Hello, World!")
	s.Pop()
	s.Top()
}
