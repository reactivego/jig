package templ

import (
	"regexp"
)

// Generic is a struct containing information about a generic type that can be
// used with concrete type information to generate a specialized type.
type Generic struct {
	// Packagename is the name of the package in which the generic was found.
	// e.g. "rx"
	PackageName string
	// Name of the generic uniquely identifies it.
	// e.g. "Observable<Foo> Map<Bar>"
	Name string
	// Vars contains the template vars found in the generic.
	// e.g. ["Foo", "Bar"]
	Vars []string
	// Needs contains the generics this one requires
	// e.g. ["Observable<Foo>", "<Foo>Observer"]
	Needs []string
	// Embeds contains the names of the generics that match the types this type
	// embeds using golang type embedding. Only meant for single type generics
	// (e.g. "Observable<Foo>") not for "Name Method" combinations.
	// This helps jig to specialize methods of embedded types to satisfy dependencies
	// on the embedding type. Adds the generics for the embedded type's methods to
	// the set of generics that can be specialized for this type.
	// e.g. ["Observable<Foo>"]
	Embeds []string

	// RequiredVars contains the set of template vars that are not allowed to be unassigned.
	// e.g. ["Foo"]
	RequiredVars []string

	// identifier is the generic name with all spaces and angle brackets around
	// the template variable names removed.
	// e.g. "ObservableFoo_MapBar"
	identifier string
	// signature is a regular expression to which type signatures are atempted to be
	// matched in order to find out if this generic is compatible with the type signature.
	// e.g. ^Observable([[:word:]]+) Map([[:word:]]+)$
	signature *regexp.Regexp
}

func (t Generic) nameID() string {
	return "N" + t.identifier
}

func (t Generic) sourceID() string {
	return "S" + t.identifier
}

// PackageWriter is the interface expected by the specializer to add
// generated source fragments to the package.
type PackageWriter interface {
	// Typemap is used to tell the specializer about the real type to use instead of the display type.
	// Usefull for generaring code for non-exported types.
	// e.g. display type MyFoo maps to real type myfoo struct { ... }
	// The display type is used in signatures e.g. NewMyFooDict() but the actual type definitions
	// need to know about the real type myfoo
	Typemap() map[string]string

	// HasGeneratedSource will take a fragment name and return true if this has
	// already been generated as part of the package.
	HasGeneratedSource(name string) bool

	// GenerateSource given a package name, a fragment name and the source content
	// for a fragment will generate and append the source to the package, returning
	// an error if something goes wrong.
	GenerateSource(packageName, name, source string) error
}

// Specializer is used during the generics definition phase to Add generics while
// they are found in the source of the package being processed (or any packages
// imported by that package). Then after that is finished, Sort is used to sort
// all generics once.
// During the type-checking there is an algorithm that detects missing types. This algorithm
// will then use GenerateCodeForType() to find a matching generic and then generate code that
// implements the missing type using the passed in PackageWriter interface.
type Specializer interface {
	// Add will take the source and create a template based on it.
	Add(t Generic, source string) error

	// After adding all generics call Sort() once to sort all generics from longest to shortest Name length.
	// This will make longest (and therefore more precise) matches match first.
	Sort()

	// GenerateCodeForType will specialize generics to implement the undefined type, function,
	// field or method specified in the signature param. And error is only returned if something
	// unexpected went wrong during the generating process. If no code can be generated for a
	// signature, that is not considered an error.
	GenerateCodeForType(pkg PackageWriter, signature string) ([]string, error)
}
