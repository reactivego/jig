package templ

import (
	"fmt"
	"strings"
)

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
				messages = []string{fmt.Sprintf("generating %q", signature), "  " + msg}
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

// apply contains all info needed to call ExecuteTemplate on a go template
// in order to apply the template with name Template.Name to the data passed
// in Dot.
type apply struct {
	*Template
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
func (tpls *templatemanager) find(signature string, types []string) (*Template, []string) {
	for _, t := range tpls.Templates {
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
