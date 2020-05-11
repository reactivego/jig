package stack

import _ "github.com/reactivego/jig/example/stack/generic"

// Example that excercises the stack causing jig to generate the Stack type and all
// its methods we want to export into the stack.go file.
func Example() {
	var s Stack
	s.Push(42)
	s.Pop()
	s.Top()
}
