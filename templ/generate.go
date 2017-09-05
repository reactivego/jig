package templ

import (
	"fmt"
	"strings"
)

func (tpls *templatemanager) GenerateCodeForType(pkg PackageWriter, signature string) (messages []string, err error) {
	var applies []*apply
	err = generateApplies(tpls, signature, func(a *apply) {
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
func generateApplies(tpls *templatemanager, signature string, next func(*apply)) error {
	var (
		known    = make(map[string]struct{})
		generate func(signature string) error
	)
	generate = func(signature string) error {
		// we're generating for signature, so mark it as known to break any reference cycles.
		known[signature] = struct{}{}

		// tpl is the template that matches signature e.g. "ObservableInt32 MapFloat64" and types
		// contain the types in the signature e.g. ["Int32","Float64"]
		tpl, types := tpls.find(signature)
		if tpl != nil {
			if len(tpl.Vars) != len(types) {
				return fmt.Errorf("signature %q does not match template %q", signature, tpl.Name)
			}
			for _, need := range tpl.Needs {
				// Convert need of the form e.g. Observable<Foo> into ObservableInt32 assuming Foo == "Int32"
				for i, varname := range tpl.Vars {
					need = strings.Replace(need, fmt.Sprintf("<%s>", varname), types[i], -1)
				}
				// Check if this need is known
				if _, present := known[need]; !present {
					err := generate(need)
					if err != nil {
						return err
					}
				}
			}
			next(&apply{tpl, types})
			return nil
		}

		// When we get here we did not match the signature to any template.
		// Find template with just the name of the type and see if it embeds other types.
		// Create alternative signatures by combinging embeded types and original method part.

		//  Looked for e.g. for signature "ConnectableInt SubscribeOn" doesn't exist because no template for this.
		if fields := strings.Fields(signature); len(fields) == 2 {
			name, method := fields[0], fields[1]
			//  Lookup "ConnectableInt" by itself and see if embeds other types, yes "Observable<Foo>"
			tpl, types := tpls.find(name)
			if tpl != nil && len(tpl.Embeds) > 0 {
				if len(tpl.Vars) != len(types) {
					return fmt.Errorf("signature %q does not match template %q", name, tpl.Name)
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
						err := generate(embed)
						if err != nil {
							return err
						}
					}
				}
			}
		}

		// No need to return an error, the type checking loop will eventually report missing type to the user.
		return nil
	}
	return generate(signature)
}

func (tpls *templatemanager) find(signature string) (*Template, []string) {
	for _, t := range tpls.Templates {
		if t.signature != nil {
			sigmatch := t.signature.FindStringSubmatch(signature)
			if len(sigmatch) == 0 {
				continue
			}
			return t, sigmatch[1:]
		}
	}
	return nil, nil
}
