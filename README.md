# JIG - Just-In-time Generics for Go

According to **Wikipedia** a jig is a type of custom-made tool used to control the location and/or motion of parts or other tools. An example of a jig is when a key is duplicated; the original is used as a jig so the new key can have the same path as the old one.

What we propose is adding generics to Go by using an example based approach. A jig is a working piece of code that is written in terms of a place-holder (metasyntactic) type.

Let's look at an extremely simple example, a stack with just a single Push method defined:

```go
package stack

type foo int

type FooStack []foo

func (s *FooStack) Push(v foo) {
	*s = append(*s, v)
}
```

We see a couple of things here.
1. A metasyntactic place-holder type foo is introduced.
2. The definition of the FooStack in terms of the foo type.
3. The definition of the Push method in terms of the foo type.

This code will compile, and you can write tests to validate the code.

To generate code we use a command-line tool appropriately named jig. Jig uses the example code as a template for generating specializations where the place-holder type is replaced by another type.

Now in order for jig to know what code should be used as templates, we need to add comment pragmas (jig:template) to the code.

```go
package stack

type foo int

//jig:template <Foo>Stack

type FooStack []foo

//jig:template <Foo>Stack Push

func (s *FooStack) Push(v foo) {
	*s = append(*s, v)
}
```

Again, we see a couple of things here.
1. Two templates are declared: type **&lt;Foo&gt;Stack** and method **&lt;Foo&gt;Stack Push**.
2. There must be no space in front of the jig:template comment pragma declaration.

This is enough for jig to work with. Now we can create a little program that uses this stack:

```go
package main

import (
	_ "github.com/reactivego/jig/example/stack"
)

//jig:file stack.go

func main() {
	var s StringStack
	s.Push("Hello, World!")
}
```

Things to note are:
1. We are importing the stack template package purely to have it in our context as we use _ as the name.
2. The jig:file pragma tells jig to store generated code in a file named stack.go
3. Create a variable s of type StringStack to indicate we want a stack of strings.
4. Jig knows all builtin types and knows that String really means the actual type string.

As you can see, we are interested in a StringStack. After running jig. The following code is found in the file stack.go.

```go
package main

//jig:name StringStack

type StringStack []string

//jig:name StringStackPush

func (s *StringStack) Push(v string) {
	*s = append(*s, v)
}
```

Things to note are:
1. jig:name uniquely identifies a fragment of generated code, StringStack and StringStackPush respectively.
2. All occurrences of Foo in the template have been replaced with String.
3. All occurrences of foo in the template have been replaced with string.

Now when we run ```go run main.go stack.go``` the program will be build and run correctly.

Jig's unique selling points are:

1. Compilation driven code generation, generates only code to implement missing types.
2. Templates are working Go code that can be tested and build.
3. A template library is a normal Go gettable package.

Because code is only generated when needed, there is no explosion of specialized code for templates that combine multiple type variables. Methods and types that are not directly or indrectly used by the user will never cause code to be generated.
