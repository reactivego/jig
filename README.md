# jig

    $ go get github.com/reactivego/jig

[![Go Reference](https://pkg.go.dev/badge/github.com/reactivego/jig.svg)](https://pkg.go.dev/github.com/reactivego/jig#section-documentation)

`jig` is a tool for generating source code from a generics library.

## Table of Contents

<!-- MarkdownTOC -->

- [Install](#install)
- [Getting Started](#getting-started)
	- [Mixing Data Types](#mixing-data-types)
- [How does `jig` work?](#how-does-jig-work) 
	- [What is a jig?](#whats-a-jig) 
- [Usage](#usage)
	- [Command-Line](#command-line)
	- [Writing Generics](#writing-generics) 
	- [Using Generics](#using-generics)
	- [Directory Structure](#directory-structure)
- [Pragma Reference](#pragma-reference)
	- [Template Pragmas](#template-pragmas)
		- [jig:template](#jigtemplate)
		- [jig:common](#jigcommon)
		- [jig:needs](#jigneeds)
		- [jig:embeds](#jigembeds)
		- [jig:required-vars](#jigrequired-vars)
		- [jig:end](#jigend)
	- [Generator Pragmas](#generator-pragmas)
		- [jig:file](#jigfile)
		- [jig:type](#jigtype)
		- [jig:force-common-code-generation](#jigforce-common-code-generation)
		- [jig:name](#jigname)
- [Advanced Topics](#advanced-topics)
	- [Using jig inside a Template Library Package](#using-jig-inside-a-template-library-package)
	- [Type Signature Matching](#type-signature-matching)
	- [Revision Handling](#revision-handling)
	- [First and Higher Order Types](#first-and-higher-order-types)
- [Available Generics Libraries](#available-generics-libraries)
- [Acknowledgements](#acknowledgements)
- [License](#license)

<!-- /MarkdownTOC -->

## Install

```bash
$ go get github.com/reactivego/jig
```
```bash
$ jig -h
Usage of jig [flags] [<dir>]:
  -c, --clean     Remove files generated by jig
  -m, --missing   Only generate code that is missing
  -n, --nodoc     No documentation in generated files
  -r, --regen     Force regeneration of all code by jig (default)
  -v, --verbose   Print details of what jig is doing
```
## Getting Started

To get started with *jig*, let's implement our first generic library; a very simple stack.

First setup a working directory:
```bash
$ mkdir -p ~/go/hellojig/generic
$ cd ~/go/hellojig
$ go mod init hellojig
go: creating new go.mod: module hellojig
$ cd generic
$ subl stack.go
```
We're using SublimeText for editing code, hence the `subl` command.

Now (in the editor of your choice) define the `foo` type to be used as a type place-holder. Then, write three `jig:template` code fragments that use this `foo` type as a placeholder.
```go
package stack

type foo int

//jig:template <Foo>Stack

type FooStack []foo

//jig:template <Foo>Stack Push

func (s *FooStack) Push(v foo) {
  *s = append(*s, v)
}

//jig:template <Foo>Stack Pop

func (s *FooStack) Pop() (result foo) {
	slen := len(*s)
	if slen != 0 {
		result = (*s)[slen - 1]
		*s = (*s)[:slen - 1]
	}
	return
}
```
Build the code to verify that it is correct.
```bash
$ go build 
$ 
```
Next step is writing a little program that uses the library:
```bash
$ cd ..
$ subl main.go
```
Now enter the following code:
```go
package main

import _ "hellojig/generic"

func main() {
  var stack StringStack
  stack.Push("Hello, World!")
  println(stack.Pop())
}
```
Run the code with the *go* tool. Expect it to *fail*:
```bash
$ go run *.go
# command-line-arguments
./main.go:6:13: undefined: StringStack
$ 
```
Run the *jig* tool to generate code:
```bash
$ jig -v
found 3 templates in package "stack" (hellojig/generic)
generating "StringStack"
  StringStack
generating "StringStack Push"
  StringStack Push
generating "StringStack Pop"
  StringStack Pop
writing file "stack.go"
```
Run the code again with the *go* tool. Now it should *succeed*:
```bash
go run *.go
Hello, World!
$ 
```
The *jig* tool generated the code and wrote it to the file `stack.go`.
```go
// Code generated by jig; DO NOT EDIT.

//go:generate jig

package main

//jig:name StringStack

type StringStack []string

//jig:name StringStackPush

func (s *StringStack) Push(v string) {
	*s = append(*s, v)
}

//jig:name StringStackPop

func (s *StringStack) Pop() (result string) {
	slen := len(*s)
	if slen != 0 {
		result = (*s)[slen-1]
		*s = (*s)[:slen-1]
	}
	return
}
```
The generated code is a statically typed stack generated by replacing `foo` with `string` where appropriate.


### Mixing Data Types

I've read somewhere that Go doesn't need generics because it already has [`interface{}`](https://tour.golang.org/methods/14). So if you write code in terms of `interface{}`, than that code is generic right? The answer is "yes and no". Yes, because using `interface{}` allows you to use different types with the same implementation, but in the process you'll also lose something important: *static type checking*.

There are valid reasons for using `interface{}` though, and that has to do with storing values of different types in the same container. So if you want to store e.g. `int`, `string` and `struct` values in the same `slice` or `map`, then `interface{}` is what you want to use. The cost, of course, is that you then need to type assert the values you retrieve back into the underlying type before you can use them.

Another reason you may want to consider using code expressed in terms of `interface{}` in your templates, is if the implementation is very large but not type specific and you are instantiating it a lot for different concrete types. In this case a strongly typed implementation could be implemented by wrapping around an implementation defined in terms of `interface{}`. In doing so, retaining strong type checks at the cost of an extra indirection through `interface{}` based code.

By the way, *jig* is fully capable of specializing templates in terms of `interface{}` as opposed to specific types. Using `Stack` instead of `StringStack` we get:
```go
package main

import _ "hellojig/generic"

func main() {
  var stack Stack
  stack.Push("Hello, World!")
  println(stack.Pop())
}
```
*Jig* will then actually generate a stack where `foo` is replaced by `interface{}`. *Jig* recognizes the absense of a reference type name and interprets that as instruction to use `interface{}` as the replacement type.
```go
// Code generated by jig; DO NOT EDIT.

//go:generate jig

package main

//jig:name Stack

type Stack []interface{}

//jig:name StackPush

func (s *Stack) Push(v interface{}) {
	*s = append(*s, v)
}

//jig:name StackPop

func (s *Stack) Pop() (result interface{}) {
	slen := len(*s)
	if slen != 0 {
		result = (*s)[slen-1]
		*s = (*s)[:slen-1]
	}
	return
}
```
## How does `jig` work?
*Jig* works by repeatedly compiling your code. Every compile cycle it may get back several errors that tell it about missing types or missing methods. *Jig* then creates type signatures for these errors and then goes through all the templates it knows of to see if it can specialize them so they'll fix the errors. After generating all the code, *jig* will then loads the whole program into memory again and compiles it again. *Jig* knows when to stop, when either no more errors were found or when no code could be generated to fix an error. At that time remaining errors are reported to the screen.

*Jig* generates the minimal amount of code needed to make your code compile. So even for a huge generics library, only the code that is needed is actually generated into your package. I jokingly call this approach _**j**ust-**i**n-time-**g**enerics_ (jig) for Go.

*Jig* has some unique selling points:

- To use generics, no explicit specialization of templates is needed. Types are automatically discovered by `jig`.
- Templates can be specialized to a specific type (e.g. `int32`, `string`, `Employee`) **or** to the *any* type (i.e. `interface{}`).
- Compilation drives code generation to minimize the amount of code that is generated.
- Generics templates are __working__ Go code that can be tested and build.
- A template library is a normal Go gettable package.

### What's a jig?
> A jig holds a work in a fixed location and guides a tool to manufacture a product. A jig's primary purpose therefore, is to provide repeatability, accuracy, and interchangeability in the manufacturing process. [Wikipedia](https://en.wikipedia.org/wiki/Jig_(tool)).

For this project, a *jig* is defined as a working piece of code that is written in terms of place-holder types and that is then used to generate code for specific (combinations of) types.

```go
//jig:template <Foo>Stack Push

func (s *FooStack) Push(v foo) {
    *s = append(*s, v)
}
```
> Look for the place-holders `Foo` and `foo` in the code above.

## Usage

In this section we'll be revisiting the [Getting Started](#getting-started) example but with a lot more details on what's going on behind the scenes.

### Command-Line

To get help, call jig with the `--help` or `-h` flag:

```bash
jig -h
```
```bash
$ jig -h
Usage of jig [flags] [<dir>]:
  -c, --clean     Remove files generated by jig
  -m, --missing   Only generate code that is missing
  -n, --nodoc     No documentation in generated files
  -r, --regen     Force regeneration of all code by jig (default)
  -v, --verbose   Print details of what jig is doing
```

The *jig* command is a self-contained single binary file. When you are working on a program you open a terminal and change to the directory of the code you are developing. Running *jig* without any parameters will remove any previously generated code and it will then generate all code fresh.

You can run *jig* with the `--missing` or `-m` flag to only find out what types are missing and then generate and add this new code to the already exisiting code.

By default *jig* is quiet unless it finds an error. To make *jig* more chatty use the `--verbose` or `-v` flag. 

The generics *jig* uses are picked up from the packages that are imported by your code. So if your code is not importing a library, then *jig* will not be able to find it. So it is not enough to use `go get <generics library>` to install the library in your `GOPATH`, you will also need to `import _ "<generics library>"` it in your code. To see *what* generics *jig* is finding and *where*, run it like this:

```bash
$ jig -v
removing file "stack.go"
found 4 jig templates in package "github.com/reactivego/jig/example/stack/generic" 
...
```

The code generated by *jig* contains the line `//go:generate jig`. This will allow you to run for example `go generate ./...` to regenerate all files generated by jig.

### Writing Generics

Jig templates should be fairly granular. Every type and every method associated with that type and every function is normally stored in its own `jig:template`. The reason for that, is that methods on which your code does not depend will be left out of the generated code. Also if the templates are not granular enough *jig* may e.g. fail to find a template for a missing method. In practice these issues are quickly detected and fixed during template library development while you write tests for your library and *jig* reports errors about things it can't generate.

The approach we take with *jig* is to have generic code written in terms of place-holder types like `Foo` and `Bar`. These are called [metasyntactic](https://en.wikipedia.org/wiki/Metasyntactic_variable) type names. So using these type names you can write normal Go code that will compile when you `go build` it. Let's analyze the simplified example of a stack with just a `Push` and `Pop` method:

```go
package stack

type foo int

type FooStack []foo

func (s *FooStack) Push(v foo) {
	*s = append(*s, v)
}

func (s *FooStack) Pop() (foo, bool) {
	var zeroFoo foo
	if len(*s) == 0 {
		return zeroFoo, false
	}
	i := len(*s) - 1
	v := (*s)[i]
	*s = (*s)[:i]
	return v, true
}
```
> *NOTE*
> - This is strongly typed Go code in terms of place-holder type `foo`.
> - `FooStack`, `(FooStack) Push` and `(FooStack) Pop` have `Foo` in somewhere in their name.
> - You're not limited to using `Foo`, you can use any place-holder name you want, really.
> - The use of both `Foo` and `foo`. Where `Foo` refers to a type and `foo` is the actual name of the type.

This code will compile, and you can write tests (in a different directory) to validate the code. Jig will use the code as a template for generating specializations where the place-holder type is replaced by another type, but also the use of those place-holders names in identifiers and even comments will be replaced with proper type names by *jig*.

However, in order to operate correctly, jig needs some help in determining what's part of a template and what's just support code. For that we use comment pragmas to tell jig what's what. To tell jig about templates we use `jig:template`. Below, we've added these pragmas to the generic stack code:

```go
package stack

type foo int

//jig:template <Foo>Stack

type FooStack []foo

//jig:template <Foo>Stack Push

func (s *FooStack) Push(v foo) {
	*s = append(*s, v)
}

//jig:template <Foo>Stack Pop

func (s *FooStack) Pop() (foo, bool) {
	var zeroFoo foo
	if len(*s) == 0 {
		return zeroFoo, false
	}
	i := len(*s) - 1
	v := (*s)[i]
	*s = (*s)[:i]
	return v, true
}
```
> *NOTE*
> - Three templates are declared: `<Foo>Stack`, `<Foo>Stack Push` and `<Foo>Stack Pop`.
> - See how we use `<` and `>` to tell *jig* what part of an identifier is a type var.
> - Comment pragmas cannot have space between `//` and `jig:template`.
> - A comment pragma must be separated from the actual code by a an empty line.

The information provide by these three pragmas is enough for jig to work with. This simple stack is on github as part of the jig example code. To use it `import _ "github.com/reactivego/jig/example/stack/generic"`.

### Using Generics
Now let's create a little program that uses this generic stack:

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
> *NOTE*
> - We're importing `github.com/reactivego/jig/example/stack/generic` only so jig can access it.
> - The `jig:file` pragma tells jig to store generated code in a file named `stack.go`
> - We create a `StringStack` variable to indicate we want a stack of strings.
> - *jig* knows all built-in types and knows that `String` really means the actual type `string`.

After dropping to the command-line and changing to the directory where this file is located, we run jig.

`jig`

If everything went well you will see no output message, but the following code is now present in the file `stack.go`.

```go
// Code generated by jig; DO NOT EDIT.

//go:generate jig

package main

//jig:name StringStack

type StringStack []string

//jig:name StringStackPush

func (s *StringStack) Push(v string) {
	*s = append(*s, v)
}
```
> *NOTE*
> - There is a `go:generate` comment pragma to run `jig`.
> - The generated `jig:name` pragmas uniquely identify every fragment of generated code.
> - All occurrences of `Foo` and `foo` from the templates have been replaced with `String` and `string`.
> - There is no implementation of `Pop` in the generated code, because it isn't needed!

When we run `go run main.go stack.go` the program will be build and run correctly.

The next thing we'll do is use `Stack` instead of `StringStack`, to see what happens.

```go
package main

import (
	"fmt"
	_ "github.com/reactivego/jig/example/stack/generic"
)

//jig:file stack.go

func main() {
	var s Stack
	s.Push("Hello, World!")
	if value, ok := s.Pop(); ok {
		fmt.Println(value)
	}
}
```
> *NOTE*
> - We have also added a call to `Pop` and `fmt.Println`.

Now calling `jig -v` will regenerate the `stack.go` file.

```bash
$ jig -v
removing file "stack.go"
found 4 jig templates in package "github.com/reactivego/jig/example/stack/generic"
generating "Stack"
  Stack
generating "Stack Push"
  Stack Push
generating "Stack Pop"
  Stack Pop
writing file "stack.go"
```

Following is what was generated:
```go
// Code generated by jig; DO NOT EDIT.

//go:generate jig

package main

//jig:name Stack

type Stack []interface{}

var zero interface{}

//jig:name StackPush

func (s *Stack) Push(v interface{}) {
	*s = append(*s, v)
}

//jig:name StackPop

func (s *Stack) Pop() (interface{}, bool) {
	if len(*s) == 0 {
		return zero, false
	}
	i := len(*s) - 1
	v := (*s)[i]
	*s = (*s)[:i]
	return v, true
}
```
> *NOTE*
> - The code is now generated in terms of `interface{}`.
> - There is now an implementation of `(Stack) Pop` present.

If we now run `go run main.go stack.go` we get the following output:

```bash
$ go run main.go stack.go
Hello, World!
```

Switching back to a type safe `StringStack` is a matter of simply using that instead of `Stack` and running `jig` again to specialize stack for the `string` type.

If you've made it to here, you will probably be able to become productive with *jig* right away, either as a user or hopefully also as an author of just-in-time generics libraries.

### Directory Structure

While developing *jig* we discovered some best practices around writing a generic Go library. There is an effective way of structuring it and we'll introduce that here. Let's look at an example library `github.com/reactivego/jig/example/stack`. It's located in our *jig* repository on GitHub.

```
$GOPATH/src/github.com/reactivego/jig/example/stack
                                              ├── generic
                                              └── test
                                                  ├── Pop
                                                  ├── Push
                                                  └── Top
```
We see two levels here; `stack` at the root level, `generic` and `test` below that. The root directory `stack` exports our library specialized on the type `interface{}`. The `generic` directory contains the actual generics library and the `test` directory contains tests code that exercises the generics library. For every exported function and method there is a separate directory. Code is generated by *jig* in the `stack` directory and in the `test` directories.

##### stack

From the `stack` directory, export a package named `stack`. This package is generated from generics in directory `generic` specialized on the type `interface{}`. This makes the package useful without users actually needing to run **jig** themselves.

The exported package is heterogeneous, because it uses only `interface{}` values. So, you can call functions and methods using values of any type. The consequence of this, is that the code is not type-safe.

To make *jig* generate code into this package, we need to actually use the generics somehow. This is accomplished by writing examples and placing them in `example_test.go`. These examples are written using generics from the `generic` directory. Then run *jig* in the root `stack` directory to generate the code into the `stack.go` file.

Any documentation of the package should go into a separate `doc.go` file, because on godoc.org documentation from `example_test.go` is not show.

```
$GOPATH/src/github.com/reactivego/jig/example/stack
                                              ├── doc.go
                                              ├── example_test.go
                                              └── stack.go
```

##### generic

The `generic` directory is the actual generics library. This directory should contain all generic templates. It should be code that can be compiled normally with Go.

One thing to note is that the package name is actually `stack`, although the code is in the directory `generic`. This is needed so *jig* knows what name to use for the files that it generates.

There should be no tests here, just the generic templates.

```
$GOPATH/src/github.com/reactivego/jig/example/stack
                                              ├── generic
                                              │   └── stack.go
                                              ├── doc.go
                                              ├── example_test.go
                                              └── stack.go
```

##### test

The `test` directory contains test code for testing all aspects of the generics library. The code here should exercise the generics library using different types. Every function or method exported by the library gets its own directory. This allows testing of functionality in isolation and works great with how godoc.org presents the documentation.

Running *jig* in the `test` sub-directory will generate code into a `stack.go` file. 

```
$GOPATH/src/github.com/reactivego/jig/example/stack
                                              ├── generic
                                              │   └── stack.go
                                              ├── test
                                              │   ├── Pop
                                              │   │   ├── doc.go
                                              │   │   ├── example_test.go
                                              │   │   └── stack.go
                                              │   └── doc.go
                                              ├── doc.go
                                              ├── example_test.go
                                              └── stack.go
```
## Pragma Reference

Pragmas are the means by which you express your intent to *jig*. There are two different kinds of pragmas. The first kind are called Template Pragmas, they are used in your template library code to inform *jig* about templates and how templates relate to each other. The second kind are called Generator Pragmas and are used in the program that uses the template library.

Note: pragmas **MUST** always be separated from the template code by an **empty line**; otherwise they will be seen as comment blocks that belong to the code.

### Template Pragmas
Tell *jig* about your templates. Use these pragmas when you are **writing** a template library.

#### jig:template

Use this to define a template.

The pragma `jig:template` defines the start of a new template. Consequently, it also marks the end of any previous template. Following are examples of this pragma:

```go
//jig:template Subscriber

//jig:template Observable<Foo>

//jig:template Observable<Foo> Map<Bar>
```

The first one defines a template without any template variables. The second one defines a template for the type `ObservableFoo` with a single template var `Foo`. The third one defines a template for method `MapBar` on `ObservableFoo` with two template vars `Foo` and `Bar`.

The code that follows a `jig:template` pragma is the actual jig. Any occurence of `Foo` and `Bar` when specialized for conrete types will (during code generation) be replaced with a capitalized type name e.g. `Int32`, `String`, whereas any occurence of `foo` and `bar` will be replaced with an actual type name e.g. `int32`, `string`.

Following is a real example of a method jig in the `github.com/reactivego/rx/generic` template library.
```go
//jig:template Observable<Foo> Map<Bar>

// MapBar transforms the items emitted by an ObservableFoo by applying a function to each item.
func (o ObservableFoo) MapBar(project func(foo) bar) ObservableBar {
	observable := func(observe BarObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {
		observer := func(next foo, err error, done bool) {
			var mapped bar
			if !done {
				mapped = project(next)
			}
			observe(mapped, err, done)
		}
		o(observer, subscribeOn, subscriber)
	}
	return observable
}
```
#### jig:common
Use this pragma **sparingly**.

The pragma `jig:common` marks this template as needed by **practically every** template. You should mark templates as common that are **always** needed in generated code. This is an alternative to having to put a `jig:needs` pragma in every template. Templates that are marked as common must not have any template vars. Common code is always generated into the program that is using the template library and must therefore really be **common** code.

There are generally just a few templates marked as `jig:common` because there is normally only a small amount of common support code. We want the templates inside the library to still build like normal Go code. So we refer to the shared code directly and have that same common code also written to the package that is using the templates.

#### jig:needs
Optional but handy pragma to **speed up** code generation.

Use this pragma to tell *jig* explicitly what other templates a specific template needs. This will generate the needed template first and prevent an additional template expansion iteration. Not telling *jig* the template is needed will cause the template to be found when code fails to compile and *jig* has to trace back to see what templates to expand in order to fix the compilation error. Following is an example of a needs pragma:

	//jig:needs Observable<Foo> Concat, SubscribeOptions

As mentioned, the use of this pragma can speed-up code generation considerably, because *jig* will need less (time consuming) compile iterations. Types that would be found missing in the next compile will already have been generated. 

In corner cases, *jig* also decides between two equally viable matches based on the template variables already discovered. So `jig:needs` clauses can sometimes influence which template is used in a subtle way.

Note, you can't specify a needed template that is only a partial match. Say you need a type e.g. `TimeStamp<Foo>Observer`. Unless there is an actual template `TimeStamp<Foo>Observer` it will be reported as 'missing' when jig tries to generate code for it. This is something typically encountered when you introduce a new type in your template definition. What you actually would like to specify is `<TimeStamp<Foo>>Observer`, but that is not supported yet by jig.

#### jig:embeds
The pragma `jig:embeds` can be used to tell jig that a certain type embeds other types. Generating a method for an embedded type could satisfy a missing method that is needed but that the curent type does not specify a template for. Code generation will then specialize the template code for the embedded type.

	//jig:embeds Observable<Foo>

Following is a real example of a jig that uses the `jig:embeds` pragma in the `github.com/reactivego/rx/generic` template ibrary.

```go
//jig:template Connectable<Foo>
//jig:embeds Observable<Foo>
//jig:needs SubscribeOptions

// ConnectableFoo is an ObservableFoo that has an additional method Connect()
// used to Subscribe to the parent observable and then multicasting values to
// all subscribers of ConnectableFoo.
type ConnectableFoo struct {
	ObservableFoo
	connect func(options []SubscribeOptionSetter) Subscriber
}
```
#### jig:required-vars

Use this **seldomly** to prevent a template specializing on `interface{}`. A case in which you'd want to use this is when a specific version specialized on `interface{}` is already available.

The pragma `jig:required-vars` contains the list of template vars that have to be assigned a **non empty** value in order for a template to match a specific type signature. Requiring a template var to be non empty will prevent the template from matching to the any type and prevents specializing the template for `interface{}` for that template var.

Following is an example of `jig:required-vars`:
```go
//jig:required-vars Foo, Bar
```
Required vars specifies that type names must be something like `String` or `Int` and can't be empty. So e.g. for template `Stack<Foo>`, a user writing code like `StackString` and `StackInt` is fine, but writing just `Stack` suggesting the use of `interface{}` is not.

#### jig:end
The pragma `jig:end` explicitly marks the end of a template.

### Generator Pragmas
These pragmas should be put into code **using** a template library. They tell *jig* how to change the way in which it generates code.

#### jig:file

Tell *jig* where to put generated code.

The pragma `jig:file` tells *jig* how to name the files it is generating. *Jig* can either generate one big file containing all code; or it can generate separate files. You put this pragma in a single file in the package you are writing, e.g. in your `doc.go` file. 

By default when `jig:file` is not found the filename template `{{.package}}.go` is used. This creates a single file per template library. The `{{.package}}` variable is the name of the package where the generics are from converted to lowercase. So when you use templates from both `github.com/reactivego/multicast/generic` and `github.com/reactivego/rx/generic` you will find two files with generated code; `multicast.go` and `rx.go`. Every generated code fragment is added to the file named after the package the template came from. So templates originating from the same template library will end up in the same generated file.

You can customize the name of the generated files. It can optionally contain some predefined variables that get replaced with contextual information. In the name you can use the variables `{{.Package}}`, `{{.package}}`, `{{.Name}}` and `{{.name}}`. Package is the name of the package that contains the *jig* template being expanded and Name is the signature of the source fragment being expanded. Capitalized variables (e.g. `{{.Package}}`) are created via `strings.Title(var)` and lowercase variant (e.g. `{{.package}}`) are create via `strings.ToLower(var)` allowing full control over the generated filename. Examples, assuming the template package name is `rx`:

	SomeName.go              -> SomeName.go
	jig{{.Package}}.go       -> jigRx.go
	jig{{.package}}.go       -> jigrx.go
	{{.Package}}{{.Name}}.go -> RxObservableInt32.go
	{{.package}}{{.name}}.go -> rxobservableint32.go

#### jig:type

A bit of a kludge to tell *jig* about types that are not exported from your package.

You need this pragma to use templates from a template library with your own custom types like e.g. `vector`. So, types that start with a lowercase letter and are therefore internal to your package.

By default *jig* assumes type names that are not part of the language will start with a capital letter. The pragma `jig:type` allows you to specify the actual type for a capitalized type reference. 

This is best illustrated with an example...

Given a program that we are writing. We don't want to export our internal type 'vector':
```go
type vector struct{ x, y float64 }
```
But *jig* will see the type 'Vector' you used in a referenced 'Stack' template
```go
var vstack StackVector.
```
So we now need to tell jig to use 'vector' in stead of 'Vector'.
```go
//jig:type Vector vector
```

So *jig* will correctly generate code, for example:

```go
// Code generated by jig; DO NOT EDIT.
var v vector
```

So, whenever you want to generate code for an internal type that starts with a lowercase letter, you will need to use the `jig:type` pragma to tell *jig* about it.

```go
//jig:type Size size
//jig:type Points []point
```

As shown in the example, it is also possible to use punctuation e.g. `[]`, `*` in the actual type name.

#### jig:force-common-code-generation
You will probably **never** need this pragma.

This **may** be useful though when you are writing a template library. Templates must be written so that Go can compile them. If you don't, then you can't import the template library in a program that wants to use it. Often, making templates compile requires code generated from templates in the library itself. 

 There is no problem in having *jig* generate code from templates inside the library you are developing. It even automatically skips generating certain code when it detects it's operating in this 'inception' mode. Code explicitly marked with pragma `jig:common` or templates that are mentioned in a `jig:needs` pragma and that are not themselves specialized on a template variable for example. If *jig* would not do this, then there would be a duplication of symbols and the template library would no longer compile.

 The pragma `jig:force-common-code-generation` **forces** the generator to generate the code **anyway** for the excluded templates. You would normally only use this to see what templates *jig* is skipping. 

#### jig:no-doc
You will probably **never** need this pragma.

Use this to skip generating documentation for the code you are generating. This is practical if you are developing a template package and you want to see just the code without all the documentation getting in the way.

#### jig:name
You will **never** use this yourself.

The pragma `jig:name` is actually **written by _jig_** to identify code fragments it generated. When *jig* is updating your code, it reads back already generated code to see which fragments are already present and it will be able to generate only the code that is really needed.
E.g. `ObservableInt32MapFloat32` might be a name that is written out for code generated for a template named `Observable<Foo> Map<Bar>` and using types `int32` and `float32` for `Foo` and `Bar` respectively.


## Advanced Topics

In previous sections, we've been looking at the mosts commonly used aspects of *jig*. We now look at more advanced subjects that detail some of the more difficult aspects of this generic programming tool.


### Using jig inside a Template Library Package

In order for template code to compile, it depends on code that could potentially also be generated with *jig*. An example can be found in `github.com/reactivego/rx/generic` where to implement the template code for mapping from one observable to another you need to have both type `ObservableFoo` as well as type `ObservableBar` available. However, `ObservableFoo` is the template and `ObservableBar` is just an instance of that template for the `bar` type.

Fortunately, *jig* has been designed to fully support this. You just need to take some precautions by instructing the generator side of *jig* exactly what to generate.

The following code fragment shows how the `rx` library is configured for internal *jig* code generation.
```go
package rx

// foo is the first metasyntactic type. Use the jig:type pragma to tell jig that
// Foo is the reference type name for actual type foo. Needed because we're
// generating code into rx.go for foo.

//jig:type Foo foo

type foo int

// bar is the second metasyntactic type. Use the jig:type pragma to tell jig that
// Bar is the reference type name for actual type bar. Needed because we're
// generating code into rx.go for bar.

//jig:type Bar bar

type bar int
```
> *NOTE*
> - `jig:type` tells jig about unexported types `foo` and `bar`, referenced by names `Foo` and `Bar` respectively.

### Type Signature Matching

*Jig* was build by looking at the errors that Go reports when it is trying to build code.

The following error is an example of what happens when you use a type that doesn't exist:

```bash
./main.go:11:8: undefined: StringStack
```

So if you define a template with name `<Foo>Stack`, this is something *jig* can then work with, replacing occurences of `Foo` with `String` and occurences of `foo` with `string`.

The following error is an example of using a method on a type that does exist:

```bash
./main.go:12:3: s.Push undefined (type StringStack has no field or method Push)
```

So Go reports that it knows of the type, but can't find the field or method. So if you defined a template `<Foo>Stack Push`, then this again is something that *jig* can work with.

All of the detection capabilities of *jig* are build around just 5 simple regular expressions matching with errors reported by Go.

```regexp
^(.*):([0-9]*):([0-9]*): undeclared name: (.*)$
^(.*):([0-9]*):([0-9]*): invalid operation: .* [(]value of type [*](.*)[)] has no field or method (.*)
^(.*):([0-9]*):([0-9]*): invalid operation: .* [(]variable of type [*](.*)[)] has no field or method (.*)
^(.*):([0-9]*):([0-9]*): invalid operation: .* [(]value of type (.*)[)] has no field or method (.*)
^(.*):([0-9]*):([0-9]*): invalid operation: .* [(]variable of type (.*)[)] has no field or method (.*)
```
> NOTE: expressions match errors reported by `loader` package, they differ slightly from errors reported by `go`.


### Revision Handling

When writing a generic library, consider how changes to your library code should propagate to the code your users create with it. Programmers are used to think in terms of API's as a contract between a library and the code that uses it. However, for template libraries that whole idea doesn't work. This is because the library is used by copying fragments of the source of the library through *jig* instead of using a compiled version.

We suggest incorporating an exported `Revision123()` function in your code that your users should then reference in their code. Note that `123` is the number of the commit or some other number that changes whenever the library gets updated for every (even minor) change.

Then when your users update your library that now contains `Revision124()`, the program that uses this library will fail to build. This missing `Revision` symbol should then be interpreted by the library users as an invitation to call `go generate` to regenerate the code, but only after they update the reference from `Revision123()` to  `Revision124()` in their code.

### First and Higher order types

One of the stranger things I experienced was the realization that you can have higher order types that can be recursively applied to increase or decrease the order level of the type. This sounds pretty abstract, so let's look at an example:

```go
//jig:template ObservableObservable<Foo> MergeAll

// MergeAll flattens a higher order observable by merging the observables it emits.
func (o ObservableObservableFoo) MergeAll() ObservableFoo {
	observable := func(observe FooObserveFunc, subscribeOn Scheduler, subscriber Subscriber) {

	<SNIP>

	}
	return observable
}
```
> NOTE the 2nd order `ObservableObservableFoo` and the first order `ObservableFoo`

The template above also matches and works for generating templates for 3rd order `ObservableObservableObservableFoo` and 2nd order `ObservableObservableFoo`. 

It's not often that you need something like this, but if you do, it is great that *jig* supports this.

## Available Generics Libraries

### [Reactive Extensions](https://www.github.com/reactivego/rx/tree/master/generic)

```go
import _ "github.com/reactivego/rx/generic"
```

### [Multicast Channel](https://github.com/reactivego/multicast/tree/master/generic)

```go
import _ "github.com/reactivego/multicast/generic"
```

## Acknowledgements

I would not have been able to write *jig* without the excellent tooling of the *Go* project. *Jig* is build on top of `golang.org/x/tools/go/loader` and uses it to perform the compilation and error detection steps. The code generation feature uses standard *Go* `text/template`. The templates are generated and compiled on the fly from the code *jig* finds in the template libraries that are imported by the code under development. Error analysis and type signature matching is all done using the standard `regexp` package.

## License

This library is licensed under the terms of the MIT License. See [LICENSE](LICENSE) file in this repository for copyright notice and exact wording.

