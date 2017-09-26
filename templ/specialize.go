package templ

import (
	"bytes"
	"fmt"
	"strings"
)

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
