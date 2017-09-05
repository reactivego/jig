package pkg

import "regexp"

// jigTemplate represents comment pragma //jig:template
const jigTemplate = "//jig:template"

// jigNeeds represents comment pragma //jig:needs
const jigNeeds = "//jig:needs"

// jigEnd represents comment pragma //jig:end
const jigEnd = "//jig:end"

// jigName represents comment pragma //jig:name
// This pragma is written alongside generated code so we can later recognize it when reading the code back.
// E.g. ObservableInt32MapFloat32 might be a name that is written out for code generated for a
// template named "Observable<Foo> Map<Bar>" and using types int32 and float32 for Foo and Bar respectively.
const jigName = "//jig:name"

// jigSupports represents comment pragma //jig:supports
// This pragma can be used with a jig:template pragma to indicate what code is supported by this template.
// Using * as the value indicates that this template is needed by (and thus supports) all other templates.
// A list of explicit template names will indicate that the mentioned templates need this template to to
// work correctly.
const jigSupports = "//jig:supports"

// jigEmbeds represents comment pragma //jig:embeds
// Use this pragma for types that embed other types and therefore generating a method for an embedded type
// could satisfy a missing type that was detected. Generation should generate the code for the embedded type.
const jigEmbeds = "//jig:embeds"

// jigFile represents comment pragma //jig:file
// This pragma defines the template for the filenames to use when generating fragments.
// In the value you can use the variables{{.Package}}, {{.package}}, {{.Name}} and {{.name}}. Package is the name of
// the package that contains the jig template being expanded and Name is the sigature of the source fragment
// being expanded. Capitalized variables are created via strings.Title(var) and lowecase variant are create via
// strings.ToLower(var) allowing full control over the generated filename.
// Examples:
// 	jig.go				  -> jig.go
//  jig{{.Package}}.go 	  -> jigRx.go
//  {{.package}}{{.Name}} -> rxObservableInt32.go
const jigFile = "//jig:file"

var reCommentPragma = regexp.MustCompile("^(//jig:[[:word:]]+)[[:space:]]+(.+)$")

// jigType comment pragma allows specifying the real type for a display type.
// Useful for specializing generic code for unexported types. e.g. //jig:type Woot woot
// In this case "Woot" would be used in the derive type names, functions and methods
// identifiers. hereas the real type is used in parameters and variable types.
const jigType = "//jig:type"

var reJigType = regexp.MustCompile("^//jig:type[[:space:]]+([[:word:]]+)[[:space:]]+([[:word:]]+)$")
