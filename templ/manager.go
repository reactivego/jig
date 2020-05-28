package templ

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

// templatemanager manages all information about defined templates and
// has associated methods to parse, add, sort, find and specialize templates.
type templatemanager struct {
	// Generics contains the generics sorted from longest to shortest Generic.Name length used
	// by Specializer.FindApply() to match type signatures to templates.
	Generics []*Generic

	// GoTemplates is the repository for defined go templates added via Specializer.Parse()
	GoTemplates *template.Template
}

func NewSpecializer() Specializer {
	return &templatemanager{
		Generics:   nil,
		GoTemplates: template.New("templates"),
	}
}

func (tpls *templatemanager) Sort() {
	sort.Sort(sort.Reverse(byNumVarsAndLength(tpls.Generics)))
}

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

func (tpls *templatemanager) Add(t Generic, source string) error {

	// identifier is Name with all space characters and angle brackets removed
	// e.g. Observable<Int> Map<Bar> becomes ObservableIntMapBar
	t.identifier = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
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
		sig = strings.Replace(sig, fmt.Sprintf("<%s>", tplvar), "([[:word:]]*)", -1)
	}
	sig = fmt.Sprintf("^%s$", sig)
	t.signature = regexp.MustCompile(sig)

	// No duplicate templates allowed, return error if one is found.
	for _, tpl := range tpls.Generics {
		if tpl.Name == t.Name {
			return fmt.Errorf("duplicate template %q", t.Name)
		}
	}
	tpls.Generics = append(tpls.Generics, &t)
	return nil
}

// stdTypeMap contains a mapping from display type to real type.
var stdTypeMap = map[string]string{
	"":           "interface{}",
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

func (tpls *templatemanager) SkipSpecialize(pkg PackageWriter, appl *apply) (bool, error) {
	dot := tpls.Dot(pkg.Typemap(), appl.types)
	name, err := tpls.expand(appl.nameID(), dot)
	if err != nil {
		return true, err
	}
	return pkg.HasGeneratedSource(name), nil
}

// Specialize will specialize a specific template using the info passed in via apply.
// When specialization succeeds, a message describing what template has been applied
// is returned. If an error was encountered, that is returned as second return value.
func (tpls *templatemanager) Specialize(pkg PackageWriter, appl *apply) (string, error) {
	dot := tpls.Dot(pkg.Typemap(), appl.types)

	name, err := tpls.expand(appl.nameID(), dot)
	if err != nil {
		return "", err
	}
	if !pkg.HasGeneratedSource(name) {
		source, err := tpls.expand(appl.sourceID(), dot)
		if err != nil {
			return "", err
		}
		err = pkg.GenerateSource(appl.PackageName, name, source)
		if err != nil {
			return "", err
		}
		// For display, generate signature based on template vars and add to messages.
		sig := appl.Name
		for i, tplvar := range appl.Vars {
			sig = strings.Replace(sig, fmt.Sprintf("<%s>", tplvar), appl.types[i], -1)
		}
		return sig, nil
	}
	return "", nil
}

func (tpls *templatemanager) expand(id string, dot map[string]string) (string, error) {
	var buf bytes.Buffer
	err := tpls.GoTemplates.ExecuteTemplate(&buf, id, dot)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (tpls *templatemanager) GenerateCodeForType(pkg PackageWriter, signature string) (messages []string, err error) {
	var applies []*apply
	missing, err := generateApplies(tpls, signature, func(a *apply) (bool, error) {
		return tpls.SkipSpecialize(pkg, a)
	}, func(a *apply) {
		applies = append(applies, a)
	})
	if err != nil {
		return nil, err
	}
	for _, apply := range applies {
		msg, err := tpls.Specialize(pkg, apply)
		if msg != "" {
			if len(messages) == 0 {
				//messages = []string{fmt.Sprintf("generating %q", signature), "  " + msg}
				messages = []string{fmt.Sprintf("generating %q %#v", signature, apply.types), "  " + msg}
				for _, msg := range missing {
					messages = append(messages, "  "+msg)
				}
			} else {
				messages = append(messages, "  "+msg)
			}
		}
		if err != nil {
			return messages, err
		}
	}
	return messages, nil
}

// apply contains all info needed to call ExecuteGeneric on a go template
// in order to apply the template with name Generic.Name to the data passed
// in Dot.
type apply struct {
	*Generic
	types []string
}

// generateApplies will recurse down the needs tree of templates matching the signature
func generateApplies(tpls *templatemanager, signature string, skip func(*apply) (bool, error), next func(*apply)) ([]string, error) {
	var (
		known    = make(map[string]struct{})
		generate func(string, []string) ([]string, error)
	)
	generate = func(signature string, parentTypes []string) (missing []string, err error) {
		// we're generating for signature, so mark it as known to break any reference cycles.
		known[signature] = struct{}{}

		// Given a type signature e.g. "ObservableInt32 MapFloat64" then tpl is the template that matches that.
		// The types string slice contain the types in the signature e.g. ["Int32","Float64"]
		tpl, types := tpls.find(signature, parentTypes)
		if tpl != nil {
			if len(tpl.Vars) != len(types) {
				return missing, fmt.Errorf("signature %q does not match template %q", signature, tpl.Name)
			}
			application := &apply{tpl, types}
			skip, err := skip(application)
			if err == nil && !skip {
				for _, need := range tpl.Needs {
					// Convert need of the form e.g. Observable<Foo> into ObservableInt32 assuming Foo == "Int32"
					for i, varname := range tpl.Vars {
						need = strings.Replace(need, fmt.Sprintf("<%s>", varname), types[i], -1)
					}

					// Check if this need is known
					if _, present := known[need]; !present {
						msgs, err := generate(need, types)
						missing = append(missing, msgs...)
						if err != nil {
							return missing, err
						}
					}
				}
				next(application)
			}
			return missing, err
		}

		// When we get here we did not match the signature to any template.
		// Find template with just the name of the type and see if it embeds other types.
		// Create alternative signatures by combinging embeded types and original method part.

		//  Looked for e.g. for signature "ConnectableInt SubscribeOn" doesn't exist because no template for this.
		if fields := strings.Fields(signature); len(fields) == 2 {
			name, method := fields[0], fields[1]
			//  Lookup "ConnectableInt" by itself and see if embeds other types, yes "Observable<Foo>"
			tpl, types := tpls.find(name, parentTypes)
			if tpl != nil && len(tpl.Embeds) > 0 {
				if len(tpl.Vars) != len(types) {
					return missing, fmt.Errorf("signature %q does not match template %q", name, tpl.Name)
				}
				// Found e.g. ConnectableInt by itself, and it has embeded types.
				for _, embed := range tpl.Embeds {
					//  Turn embedded name "Observable<Foo>" into "ObservableInt" and add signature method part.
					for i, varname := range tpl.Vars {
						embed = strings.Replace(embed, fmt.Sprintf("<%s>", varname), types[i], -1)
					}
					// Frankenconcat into e.g. "ObservableInt SubscribeOn"
					embed = fmt.Sprintf("%s %s", embed, method)
					// Now generate for e.g. "ObservableInt SubscribeOn" if it is not already known
					if _, present := known[embed]; !present {
						msgs, err := generate(embed, types)
						missing = append(missing, msgs...)
						if err != nil {
							return missing, err
						}
					}
				}
			}
		}

		// No need to return an error, the type checking loop will eventually
		// report missing type to the user.
		missing = append(missing, fmt.Sprintf("missing %q", signature))
		//fmt.Println("MISSING", signature)
		return missing, nil
	}
	return generate(signature, nil)
}

// find matches the signature against a sorted list of templates. If types has
// entries, then the types matched from the signature must be present in the
// types list.
func (tpls *templatemanager) find(signature string, types []string) (*Generic, []string) {
	for _, t := range tpls.Generics {
		if t.signature != nil {
			sigmatch := t.signature.FindStringSubmatch(signature)
			if len(sigmatch) == 0 {
				continue
			}
			if len(types) != 0 && !contains(types, sigmatch[1:]) {
				continue
			}
			if len(t.RequiredVars) != 0 {
				if len(t.Vars) != len(sigmatch[1:]) {
					continue
				}
				vartypes := make(map[string]string)
				for i, typeval := range sigmatch[1:] {
					vartypes[t.Vars[i]] = typeval
				}
				// vartypes is now e.g. map["Foo":"Int32", "Bar":"Float64"]
				rejectMatch := false
				for _, varname := range t.RequiredVars {
					if vartypes[varname] == "" {
						fmt.Printf(">>>> %+q %+q %+q %+q\n", vartypes, sigmatch[1:], t.Vars, t.RequiredVars)
						rejectMatch = true
						break
					}
				}
				if rejectMatch {
					continue
				}
			}
			return t, sigmatch[1:]
		}
	}
	return nil, nil
}

// contains returns true when all strings in sel are also present in set.
func contains(set []string, sel []string) bool {
	has := make(map[string]bool)
	for _, e := range set {
		has[e] = true
	}
	for _, e := range sel {
		if !has[e] {
			return false
		}
	}
	return true
}
