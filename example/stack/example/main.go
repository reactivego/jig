package main

import (
	"fmt"
	_ "github.com/reactivego/jig/example/stack"
)

//jig:file stack.go

func main() {
	var s StringStack
	s.Push("Hello, World!")
	if value, ok := s.Pop(); ok {
		fmt.Println(value)
	}
}
