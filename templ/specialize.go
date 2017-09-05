package templ

import (
	"bytes"
	"fmt"
	"strings"
)

// Specialize will specialize a specific template using the info passed in via apply.
// When specialization succeeds, a message describing what template has been applied
// is returned. If an error was encountered, that is returned as second return value.
func (tpls *templatemanager) Specialize(pkg PackageWriter, appl *apply) (string, error) {
	dot := tpls.Dot(pkg.Typemap(), appl.types)
	var namebuf bytes.Buffer
	err := tpls.GoTemplates.ExecuteTemplate(&namebuf, appl.nameID(), dot)
	if err != nil {
		return "", err
	}
	name := namebuf.String()
	if !pkg.HasGeneratedSource(name) {
		var sourcebuf bytes.Buffer
		err = tpls.GoTemplates.ExecuteTemplate(&sourcebuf, appl.sourceID(), dot)
		if err != nil {
			return "", err
		}
		source := sourcebuf.String()

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
