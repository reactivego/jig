# jig - Just-in-time Generics for Go
[![](http://godoc.org/github.com/reactivego/jig?status.png)](http://godoc.org/github.com/reactivego/jig)

To install *jig*, open a terminal and type:

```bash
go get github.com/reactivego/jig
```
> NOTE: only tested on Mac

*Jig* provides generics for the Go programming language. It generates idiomatic Go code from pieces of generic code written in terms of place-holder types. *Jig* is implemented as a command `jig` you run from the command-line just like `go` itself. It is designed such that it generates the minimal amount of code for an automatically detected combination of types. I call this approach *Just-in-time Generics for Go* (or **jig**) because code is generated on demand only when it is needed and not before.

If you're more interested in a walkthrough of using *jig*, then check out the [Quick Start](https://github.com/reactivego/rx/doc/QUICKSTART.md). It demonstrates how to use *jig* to generate code for generics from the `rx` template library. This document will mainly focus on template library development.

## Table of Contents

<!-- MarkdownTOC -->

- [Rationale](#rationale)
- [Unique selling points](#unique-selling-points)
- [What is a Jig?](#what-is-a-jig)
- [Command-Line](#command-line)
- [Directory Structure](#directory-structure)
- [Defining Templates](#defining-templates)
- [Strong typing versus interface{}](#strong-typing-versus-interface)
- [Using Templates](#using-templates)
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
- [Generics Libraries](#generics-libraries)
- [Acknowledgements](#acknowledgements)
- [License](#license)

<!-- /MarkdownTOC -->

## Rationale
In the summer of 2017 when things were nice and quiet at [work](https:///simplemind.eu), I thought to myself, if there is only one thing I could add to Go, what would that be? For me the answer was generics, so I set out to find a way to add generics to Go.

For more than a year, I had been working regularly in my spare time on implementing [ReactiveX](http://reactivex.io/) for Go using idiomatic strongly typed Go code. In order to do this I needed a way to write code that is independent of any particular combination of types; I needed generics. I have to confess that I'm actually not that fond of generics. Bad experiences with gibberish error messages and massive code bloat in C++ had soured me on the paradigm.

I've tried more than a handful of generic programming and code generation solutions for Go. None of these were powerfull enough to satisfy my requirements. So I decided to take a stab at creating a generic programming solution for Go myself.

## Unique selling points
- No explicit specialization of templates is needed, types are automatically discovered by *jig*.
- Templates can be specialized to a specific type (e.g. `int32`, `string`) **or** to the *any* type (i.e. `interface{}`).
- Compilation driven code generation to minimize the amount of code generated.
- Templates are working Go code that can be tested and build.
- A template package is a normal Go gettable package.

## What is a Jig?
I actually thought of the name *jig* before I knew what it meant. After finding out though, I felt it was pretty appropriate. A [*jig*](https://en.wikipedia.org/wiki/Jig_(tool)) is a device used in the manufacturing of products to hold the work and to guide a tool. An example of a *jig* is when a key is duplicated; the original is used as a *jig* so the new key can have the same path as the old one.

For this project, a *jig* is defined as a working piece of code that is written in terms of place-holder types and that is then used to generate code for specific (combinations of) types.

```go
//jig:template <Foo>Stack Push

func (s *FooStack) Push(v foo) {
    *s = append(*s, v)
}
```
> Look for the place-holders `Foo` and `foo` in the code above.

Jig templates should be fairly granular. Every type and every method associated with that type and every function is normally stored in its own `jig:template`. The reason for that, is that methods on which your code does not depend will be left out of the generated code. Also if the templates are not granular enough *jig* may e.g. fail to find a template for a missing method. In practice these issues are quickly detected and fixed during template library development while you write tests for your library and *jig* reports errors about things it can't generate.

## Command-Line

To get help, call jig with the `--help` or `-h` flag:

```bash
jig -h
```
```bash
Mac:~ $ jig -h
Usage of jig [flags] [<dir>]:
  -c, --clean     Remove files generated by jig
  -r, --regen     Force regeneration of all code by jig
  -v, --verbose   Print details of what jig is doing
```

The *jig* command is a self-contained single binary file. When you are working on a program you open a terminal and change to the directory of the code you are developing. Running *jig* without any parameters will make *jig* find out what types are missing and then generate and add this new code to the already exisiting code.

By default *jig* is quiet unless it finds an error. To make *jig* more chatty use the `--verbose` or `-v` flag. 

You can also run *jig* with the `--regen` or `-r` flag to force it to re-generate all code it had already generated previously. This is useful for when the template library has changed, or when you are no longer using some types in your code to clean out generated code no longer needed.

The templates *jig* uses are picked up from the packages that are imported by your code. So if your code is not importing a library, then *jig* will not be able to find it. So it is not enough to use `go get <template library>` to install the library in your `GOPATH`, you will also need to `import _ "<template library>"` it in your code. To see what templates *jig* is finding and where, run `jig --regen --verbose` or more succinctly `jig -rv`. So, like this:

```bash
Mac:example $ jig --rv
removing file "stack.go"
found 4 jig templates in package "github.com/reactivego/jig/example/stack" 
...
```

*Jig* works by repeatedly compiling your code. Every compile cycle it may get back several errors that tell it about missing types or missing methods. *Jig* then creates type signatures for these errors and then goes through all the templates it knows of to see if it can specialize them so they'll fix the errors. After generating all the code, *jig* will then load the whole program into memory again and compile it again. *Jig* knows when to stop, when either no more errors were found or when no code could be generated to fix an error. At that time remaining errors are reported to the screen.

The code generated by *jig* contains the line `//go:generate jig --regen`. This will allow you to run for example `go generate ./...` to regenerate all files generated by jig.

## Directory Structure

Best practice when writing a template library, is to have it separate from any commands or tests that uses it. In the directory structure below, `stack.go` in the `stack` folder contains the templates. The `example` folder contains a command `main.go` that uses the library and `stack.go` in the `example` folder contains the code generated by *jig*. For the `test` folder this is the same; `stack_test.go` contains the tests and `stack.go` the generated code.

```bash
$GOPATH/src/github.com/reativego/jig/example/
                                     └── stack/
                                        ├── example/
                                        │   ├── main.go
                                        │   └── stack.go
                                        ├── test/
                                        │   ├── stack_test.go
                                        │   └── stack.go
                                        └── stack.go
```

 A template library can easily become dependent on its own templates when it also needs to be buildable as a normal Go package. That is not a problem, *jig* can be used to generate the code on which the library depends from the templates in the library itself.

## Defining Templates
The approach we take with *jig* is to have generic code written in terms of place-holder types like `Foo` and `Bar`. These are called [metasyntactic](https://en.wikipedia.org/wiki/Metasyntactic_variable) type names. So using these type names you can write normal Go code that will compile when you `go build` it. Let's look at an extremely simple example; a stack with just a `Push` and `Pop` method:

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

This code will compile, and you can write tests (in a different folder) to validate the code. Jig will use the code as a template for generating specializations where the place-holder type is replaced by another type, but also the use of those place-holders names in identifiers and even comments will be replaced with proper type names by *jig*.

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

The information provide by these three pragmas is enough for jig to work with. This simple stack is on github as part of the jig example code. To use it `import _ "github.com/reactivego/jig/example/stack"`.

## Strong typing versus interface{}

Choosing whether to use [`interface{}`](https://tour.golang.org/methods/14) isn't really about generic programming. Although, you could use `interface{}` as a place-holder type. However, in doing so you will have to give up on compile time type checking and open yourself up to a lot of potential bugs only caught at runtime.

There are valid reasons for using `interface{}` though, and that has to do with storing values of different types in the same container. So if you want to store e.g. `int`, `string` and `struct` values in the same `slice` or `map`, then `interface{}` is what you want to use. The cost, of course, is that you then need to type assert the values you retrieve back into the underlying type before you can use them.

Another reason you may want to consider using code specified in terms of `interface{}` in your templates, is if the implementation is very large but not type specific and you are instantiating it a lot for different concrete types. In this case a strongly typed implementation could be implemented by wrapping around an implementation defined in terms of `interface{}`. In doing so, retaining strong type checks at the cost of an extra indirection through `interface{}` based code.

By the way, *jig* is fully capable of specializing templates in terms of `interface{}` as opposed to specific types. Using the `Stack` example above, if you use it in the following way:

```go
var stack Int32Stack
```

This will generate a strongly typed stack that only allows you to store `int32` values.

But using it in the following way:

```go
var stack Stack
```

This will actually generate a stack in terms of the `interface{}` type. *Jig* internally maps the absense of a reference type name to indicate `interface{}` as the actual type to use.


## Using Templates
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
> - We're importing `github.com/reactivego/jig/example/stack` only so jig can access it.
> - The `jig:file` pragma tells jig to store generated code in a file named `stack.go`
> - We create a `StringStack` variable to indicate we want a stack of strings.
> - *jig* knows all built-in types and knows that `String` really means the actual type `string`.

After dropping to the command-line and changing to the directory where this file is located, we run jig.

`jig`

If everything went well you will see no output message, but the following code is now present in the file `stack.go`.

```go
// Code generated by jig; DO NOT EDIT.

//go:generate jig --regen

package main

//jig:name StringStack

type StringStack []string

//jig:name StringStackPush

func (s *StringStack) Push(v string) {
	*s = append(*s, v)
}
```
> *NOTE*
> - There is a `go:generate` comment pragma to run `jig --regen`.
> - The generated `jig:name` pragmas uniquely identify every fragment of generated code.
> - All occurrences of `Foo` and `foo` from the templates have been replaced with `String` and `string`.
> - There is no implementation of `Pop` in the generated code, because it isn't needed!

When we run `go run main.go stack.go` the program will be build and run correctly.

The next thing we'll do is use `Stack` instead of `StringStack`, to see what happens.

```go
package main

import (
	"fmt"
	_ "github.com/reactivego/jig/example/stack"
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

Now calling `jig -rv` will regenerate the `stack.go` file.

```bash
Mac:example $ jig -rv
removing file "stack.go"
found 4 jig templates in package "github.com/reactivego/jig/example/stack"
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

//go:generate jig --regen

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
Mac:example $ go run main.go stack.go
Hello, World!
```

Switching back to a type safe `StringStack` is a matter of simply using that instead of `Stack` and running `jig` again to specialize stack for the `string` type.

If you've made it to here, you will probably be able to become productive with *jig* right away, either as a user or hopefully also as an author of just-in-time generics libraries.

## Pragma Reference

Pragmas are the means by which you express your intent to *jig*. There are two different kinds of pragmas. The first kind are called Template Pragmas, they are used in your template library code to inform *jig* about templates and how templates relate to each other. The second kind are called Generator Pragmas and are used in the program that uses the template library to configure the generator of *jig*

### Template Pragmas
Tell *jig* about your templates.

#### jig:template
The pragma `jig:template` defines the start of a new template. Consequently, it also marks the end of any previous template. Following are examples of this pragma:

	//jig:template Subscriber
	//jig:template Observable<Foo>
	//jig:template Observable<Foo> Map<Bar>

The first one defines a template without any template variables. The second one defines a template for the type `ObservableFoo` with a single template var `Foo`. The third one defines a template for method `MapBar` on `ObservableFoo` with two template vars `Foo` and `Bar`.

The code that follows a `jig:template` pragma is the actual jig. Any occurence of `Foo` and `Bar` when specialized for conrete types will (during code generation) be replaced with a capitalized type name e.g. `Int32`, `String`, whereas any occurence of `foo` and `bar` will be replaced with an actual type name e.g. `int32`, `string`.

Following is a real example of a method jig in the `github.com/reactivego/rx` template ibrary.
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
The pragma `jig:common` marks this template as purely for support needed by every generated template. There are generally just a few templates marked as `jig:common` because there is normally only a small ammount of common support code. We want the templates to still build like normal Go code. So we refer to the shared code directly (without package name prefix) and have that same common code also written to the package that is using the templates.

You may wonder why (as this is common code) we don't just import the symbols from the template library? The reason we even need `jig:common` is because code inside the template library package cannot refer to code inside that same package by package name. If that was allowed by Go, it would create an import cycle. Also, I am intentionally not considering dot imports to solve this problem.

#### jig:needs
The pragma `jig:needs` indicates what other templates this template needs for its own implementation. Although needs is optional, it guides the order of code generation as needed templates are generated first. In corner cases, *jig* also decides between two equally viable matches based on the template variables already discovered. So `jig:needs` clauses can sometimes influence which template is used.

Following is an example of a needs clause:

	//jig:needs Observable<Foo> Concat, SubscribeOptions

These clauses can also speed-up code generation considerably because *jig* will need less compile iterations. Types that would be found missing in the next compile will already have been generated. 

#### jig:embeds
The pragma `jig:embeds` can be used to tell jig that a certain type embeds other types. Generating a method for an embedded type could satisfy a missing method that is needed but that the curent type does not specify a template for. Code generation will then specialize the template code for the embedded type.

	//jig:embeds Observable<Foo>

Following is a real example of a jig that uses the `jig:embeds` pragma in the `github.com/reactivego/rx` template ibrary.

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

The pragma `jig:required-vars` contains the list of template vars that have to be assigned a non empty value in order for a template to match a specific type signature. We use this seldomly. Requiring a template var to be non empty will prevent the template from matching to the any type and prevents specializing the template for `interface{}` for that template var.

Following is an example of `jig:required-vars`:
```go
//jig:required-vars Foo, Bar
```

#### jig:end
The pragma `jig:end` explicitly marks the end of a template.

### Generator Pragmas
Tell *jig* about generating code.

#### jig:file
The pragma `jig:file` tells *jig* how to name the file(s) it is generating. The name you specify can optionally contain some predefined variables that get replaced with contextual information. In the name you can use the variables `{{.Package}}`, `{{.package}}`, `{{.Name}}` and `{{.name}}`. Package is the name of the package that contains the *jig* template being expanded and Name is the signature of the source fragment being expanded. Capitalized variables are created via `strings.Title(var)` and lowercase variant are create via `strings.ToLower(var)` allowing full control over the generated filename. Examples, assuming the template package name is `rx`:

	SomeName.go					-> SomeName.go
	jig{{.Package}}.go 			-> jigRx.go
	{{.package}}{{.Name}}.go	-> rxObservableInt32.go

By default when `jig:file` is not found the filename template `jig{{.Package}}{{.Name}}.go` is used. This will cause every code fragment to be generated into its own file.

#### jig:type
The pragma `jig:type` allows you to specify the actual type for a capitalized type reference. Useful for specializing generic code for unexported types. e.g. `//jig:type Size size` In this case `Size` would be used to refer to the type in identifiers etc. and `size` is the actual type used in parameters and variable types.

	//jig:type Size size
	//jig:type Points []point

As shown in the example, it is also possible to use punctuation e.g. `[]`, `*` in the actual type name.

#### jig:force-common-code-generation
This is only needed in template packages when using *jig* to generate supplemental code to support the main templates. In that case because you're generating in the library itself, the support code is already there and generating it again would create duplicate symbols. The pragma `jig:force-common-code-generation` tells the generator to generate the commong code anyway for templates explicitly marked with pragma `jig:common` or for templates that are mentioned in a `jig:needs` pragma and that are not themselves specialized on a template variable e.g. `SubscribeOptions`.

#### jig:name
The pragma `jig:name` is actually generated by jig itself to identify generated code fragments. When read back by the generator it knows which fragments are already present and it will be able to generate only the code that is really needed.
E.g. `ObservableInt32MapFloat32` might be a name that is written out for code generated for a template named `Observable<Foo> Map<Bar>` and using types `int32` and `float32` for `Foo` and `Bar` respectively.


## Advanced Topics

In previous sections, we've been looking at the mosts commonly used aspects of *jig*. We now look at more advanced subjects that detail some of the more difficult aspects of this generic programming tool.


### Using jig inside a Template Library Package

In order for template code to compile, it depends on code that could potentially also be generated with *jig*. An example can be found in `rx` where to implement the template code for mapping from one observable to another you need to have both type `ObservableFoo` as well as type `ObservableBar` available. However `ObservableFoo` is the template and `ObservableBar` is just an instance of that template for the `bar` type.

Fortunately, *jig* has been designed to fully support this. You just need to take some precautions by instructing the generator side of *jig* exactly what to generate.

The following code fragment shows how the `rx` library is configured for internal *jig* code generation.
```go
package rx

// Code generated to make this rx package buildable should be written to support.go

//jig:file support.go

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
> - `jig:file` is used to tell jig to write code to a file with name `rx.go` (`{{.package}}` will be replaced by `rx`).
> - `jig:type` to tell jig there are 2 unexported types `foo` and `bar` that are referred to by names `Foo` and `Bar`.

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

## Generics Libraries

- [Reactive eXtensions for the Go language](https://www.github.com/reactivego/rx/)

```go
	import _ "github.com/reactivego/rx"
```

- [Multicasting, multi sender/receiver, buffered channel](https://www.github.com/reactivego/channel/)

```go
	import _ "github.com/reactivego/channel"
```

## Acknowledgements

I would not have been able to write *jig* without the excellent tooling of the *Go* project. *Jig* is build on top of `golang.org/x/tools/go/loader` and uses it to perform the compilation and error detection steps. The code generation feature uses standard *Go* `text/template`. The templates are generated and compiled on the fly from the code *jig* finds in the template libraries that are imported by the code under development. Error analysis and type signature matching is all done using the standard `regexp` package. 


## License

This library is licensed under the terms of the MIT License. See [LICENSE](LICENSE) file in this repository for copyright notice and exact wording.

