package templ

import (
	"regexp"
	"sort"
	"text/template"
)

// PackageWriter is the interface expected by the templater to add
// generated source fragments to the package.
type PackageWriter interface {
	// Typemap is used to tell the templater about the real type to use instead of the display type.
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

// Template is a struct containing information about a generic type that can be
// used with concrete type information to generate a specialized type.
type Template struct {
	// Packagename is the name of the package in which the template was found.
	// e.g. "rx"
	PackageName string
	// Name of the template uniquely identifies it.
	// e.g. "Observable<Foo> Map<Bar>"
	Name string
	// Vars contains the template vars found in the template.
	// e.g. ["Foo", "Bar"]
	Vars []string
	// Needs contains the templates this one requires
	// e.g. ["Observable<Foo>", "<Foo>Observer"]
	Needs []string
	// Embeds contains the names of the templates that match the types this type
	// embeds using golang type embedding. Only meant for single type templates
	// (e.g. "Observable<Foo>") not for "Name Method" combinations.
	// This helps jig to specialize methods of embedded types to satisfy dependencies
	// on the embedding type. Adds the templates for the embedded type's methods to
	// the set of templates that can be specialized for this type.
	// e.g. ["Observable<Foo>"]
	Embeds []string

	// identifier is the template name with all spaces and angle brackets around
	// the template variable names removed.
	// e.g. "ObservableFooMapBar"
	identifier string
	// signature is a regular expression to which type signatures are atempted to be
	// matched in order to find out if this template is compatible with the type signature.
	// e.g. ^Observable([[:word:]]+) Map([[:word:]]+)$
	signature *regexp.Regexp
}

func (t Template) nameID() string {
	return "N" + t.identifier
}

func (t Template) sourceID() string {
	return "S" + t.identifier
}

// Templater is used during the template definition phase to Add templates while
// they are found in the source of the package being processed (or any packages
// imported by that package). Then after that is finished, Sort is used to sort
// all templates once.
// During the type-checking there is an algorithm that detects missing types. This algorithm
// will then use GenerateCodeForType() to find a matching template and then generate code that
// implements the missing type using the passed in PackageWriter interface.
type Templater interface {
	// Add will take the source and create a template based on it.
	Add(t Template, source string) error

	// After adding all templates call Sort() once to sort all templates from longest to shortest Name length.
	// This will make longest (and therefore more precise) matches match first.
	Sort()

	// GenerateCodeForType will specialize templates to implement the undefined type, function,
	// field or method specified in the signature param. And error is only returned if something
	// unexpected went wrong during the generating process. If no code can be generated for a
	// signature, that is not considered an error.
	GenerateCodeForType(pkg PackageWriter, signature string) ([]string, error)
}

func NewTemplater() Templater {
	return &templatemanager{
		Templates:   nil,
		GoTemplates: template.New("templates"),
	}
}

// templatemanager manages all information about defined templates and
// has associated methods to parse, add, sort, find and specialize templates.
type templatemanager struct {
	// Templates contains the templates sorted from longest to shortest Template.Name length used
	// by Templater.FindApply() to match type signatures to templates.
	Templates []*Template

	// GoTemplates is the repository for defined go templates added via Templater.Parse()
	GoTemplates *template.Template
}

func (tpls *templatemanager) Sort() {
	sort.Sort(sort.Reverse(byNumVarsAndLength(tpls.Templates)))
}

// byNumVarsAndLength attaches sort methods to a []*template.Template instance.
// It sorts first on the number of template variables from most to least number
// of variables. Withing each tier sorts templates on length of template name,
// from short to longest.
type byNumVarsAndLength []*Template

func (a byNumVarsAndLength) Len() int {
	return len(a)
}

func (a byNumVarsAndLength) Less(i, j int) bool {
	if len(a[i].Vars) > len(a[j].Vars) {
		// Templates with more variables are considered less.
		return true
	}
	if len(a[i].Vars) < len(a[j].Vars) {
		// Templates with less variables are considered more
		return false
	}
	// Equal number of variables, then compare length of template Name.
	return len(a[i].Name) < len(a[j].Name)
}

func (a byNumVarsAndLength) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
