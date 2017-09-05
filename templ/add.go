package templ

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// stdVar contains canonical type names used in go templates.
// The metasyntactic names in template declarations (e.g. Foo, Bar, Baz, Qux)
// are mapped by order of appearance to stdVar names (e.g. T, U, V, W) that are then
// used in the definition of go templates.
//
// e.g.
//    Say we have a template declaration "Observable<Foo> Map<Bar>"
//    In the go template we would use Observable{{.T}} Map{{.U}}
//    So Foo is mapped to T and Bar is mapped to U
//    Now when we match the template to a real type say "ObservableInt32 MapFloat32"
//    We can derive a type mapping { T: Int32, U: Float32 } and use that when applying
//    the go template.
var stdVar = []string{"T", "U", "V", "W", "X", "Y", "Z"}

func (tpls *templatemanager) Add(t Template, source string) error {

	// identifier is Name with all space characters and angle brackets removed
	// e.g. Observable<Int> Map<Bar> becomes ObservableIntMapBar
	t.identifier = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, t.Name)
	// Now remove angle brackets e.g. Observable<Int>Map<Bar> -> ObservableIntMapBar
	for _, varname := range t.Vars {
		t.identifier = strings.Replace(t.identifier, fmt.Sprintf("<%s>", varname), varname, -1)
	}

	// Name template used for generated source fragment names as value in //jig:name pragma and for use in source file name.
	nametpl := t.identifier
	// e.g. "ObservableFooMapBar" -> "Observable{{.T}}Map{{.U}}"
	for i, varname := range t.Vars {
		nametpl = strings.Replace(nametpl, varname, fmt.Sprintf("{{.%s}}", stdVar[i]), -1)
	}
	_, err := tpls.GoTemplates.Parse(fmt.Sprintf("{{define %q}}%s{{end}}\n", t.nameID(), nametpl))
	if err != nil {
		return err
	}

	// Source template for generating the actual fragment source code.
	// e.g. blahFooblahfooblah blahBarblahbarblah -> blah{{.T}}blah{{.t}}blah blah{{.U}}blah{{.u}}blah
	sourcetpl := source
	for i, varname := range t.Vars {
		sourcetpl = strings.Replace(sourcetpl, varname, fmt.Sprintf("{{.%s}}", stdVar[i]), -1)
		sourcetpl = strings.Replace(sourcetpl, strings.ToLower(varname), fmt.Sprintf("{{.%s}}", strings.ToLower(stdVar[i])), -1)
	}
	_, err = tpls.GoTemplates.Parse(fmt.Sprintf("{{define %q}}%s{{end}}\n", t.sourceID(), sourcetpl))
	if err != nil {
		return err
	}

	// Convert e.g. "Observable<Foo>" into regular expression "^Observable([[:word:]]+)$"
	// Then compile this and assign to t.signature used for matching to missing type signatures.
	sig := t.Name
	for _, tplvar := range t.Vars {
		sig = strings.Replace(sig, fmt.Sprintf("<%s>", tplvar), "([[:word:]]+)", -1)
	}
	sig = fmt.Sprintf("^%s$", sig)
	t.signature = regexp.MustCompile(sig)

	// No duplicate templates allowed, return error if one is found.
	for _, tpl := range tpls.Templates {
		if tpl.Name == t.Name {
			return fmt.Errorf("duplicate template %q", t.Name)
		}
	}
	tpls.Templates = append(tpls.Templates, &t)
	return nil
}

// stdTypeMap contains a mapping from display type to real type.
var stdTypeMap = map[string]string{
	"Bool":       "bool",
	"Byte":       "byte",
	"Complex128": "complex128",
	"Complex64":  "complex64",
	"Error":      "error",
	"Float32":    "float32",
	"Float64":    "float64",
	"Int":        "int",
	"Int16":      "int16",
	"Int32":      "int32",
	"Int64":      "int64",
	"Int8":       "int8",
	"Rune":       "rune",
	"String":     "string",
	"Uint":       "uint",
	"Uint16":     "uint16",
	"Uint32":     "uint32",
	"Uint64":     "uint64",
	"Uint8":      "uint8",
	"Uintptr":    "uintptr",
}

// Dot maps canconical stdVar names used in templates to actual types to replace them with.
// So given a list of types e.g. ["Int","Mouse","Move"] and a typemap e.g. { "Mouse": "mouse" }
// return Dot e.g. { "T": "Int", "t": "int", "U": "Mouse", "u": "mouse", "V": "Move", "v": "Move"}
func (tpls *templatemanager) Dot(typemap map[string]string, types []string) map[string]string {
	d := make(map[string]string)
	for i := 0; i < len(types); i++ {
		d[stdVar[i]] = types[i]

		T := types[i]
		if t, present := typemap[T]; present {
			T = t
		} else {
			if t, present := stdTypeMap[T]; present {
				T = t
			}
		}
		d[strings.ToLower(stdVar[i])] = T
	}
	return d
}
