# JIG - Just-In-time Generics for Go

What is a Jig?

Wikipedia
	A jig is a type of custom-made tool used to control the location and/or motion of parts or other tools.
	An example of a jig is when a key is duplicated; the original is used as a jig so the new key can have the same path as the old one.

In the same spirit but in our own software oriented context a jig is a working piece of code that is written in terms
of a place-holder type. This code is then used as a template for guiding a command-line tool named jig to generate
specializations where the place-holder type is replaced by another type.

Perhaps an simple example will make it easier to understand.



TBD Example



Jig's unique selling points

1. Generation guided by build failures.
2. Minimal generated code, only code that is actually used.
3. Templates are working Go code that can be tested and build.
4. A template library is a normal Go gettable package.

Because code is only generated when needed there is no explosion of specialized code for templates
that combine multiple type parameters.
