package pkg

import (
	"fmt"
	"go/ast"
	"os"
)

// RemoveGeneratedSources will remove all files that contain
// generated source from the package dir.
func (p *Package) RemoveGeneratedSources() ([]string, error) {
	fileset := p.GeneratedFileset()

	// Remove everything from the generated map.
	for name := range p.generated {
		delete(p.generated, name)
	}

	// Remove the files in the fileset from the PkgSpec.Files slice.
	return p.RemoveFileset(fileset)
}

// RemoveFileset will remove the passed set of files from the PkgSpec files slice.
func (p *Package) RemoveFileset(fileset map[*ast.File]struct{}) (messages []string, err error) {
	for path, file := range p.fileset {
		_, present := fileset[file]
		if present {
			messages = append(messages, fmt.Sprintf("removing file %q", path))
			// Remove physical files on disk.
			if e := os.Remove(path); e != nil {
				err = e
			}
			delete(p.fileset, path)
		}
	}
	return messages, err
}
