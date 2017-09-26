package pkg

import (
	"fmt"
	"go/ast"
	"go/printer"
	"io"
	"os"
)

// WriteGeneratedSources is used to write the generated
// sources to file(s).
func (p *Package) WriteGeneratedSources() ([]string, error) {
	return p.WriteFileset(p.GeneratedFileset())
}

// WriteFileset writes the set of files that contains generated source to disk.
func (p *Package) WriteFileset(fileset map[*ast.File]struct{}) (messages []string, err error) {
	for file := range fileset {
		path := p.Filepath(file)

		// create a file to write the new package content to.
		f, err := os.Create(path)
		if err != nil {
			return messages, err
		}
		defer f.Close()

		messages = append(messages, fmt.Sprintf("writing file %q", path))

		// actually write the package content out to file.
		if e := p.WriteFile(f, file); e != nil {
			err = e
		}
	}
	return messages, err
}

// WriteFile will write the given file to the given output.
func (p *Package) WriteFile(output io.Writer, file *ast.File) error {
	return printer.Fprint(output, p.Fset, file)
}
