package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/types"
	"io"
	"os"
	"text/template"

	"golang.org/x/tools/go/loader"
)

// Package manages the package being checked for errors and where source code will be
// generated when errors indicate missing types etc.
type Package struct {
	*loader.Config

	// Dir is the directory where jig will look for the package files.
	Dir string

	// Name is the package name found in the first source file that is scanned from the package dir.
	Name string

	// generated maps source fragment name to filepath
	generated map[string]string

	// fileset maps filepath to file instance.
	fileset map[string]*ast.File

	// allPackages is populated by checking the loaded source code.
	allPackages []*loader.PackageInfo

	// filename template for where source fragments are to be generated.
	// Depending on the given template, this may be a single file or multiple files.
	filename *template.Template

	// typemap contains a mapping of display types e.g. Foo to real types e.g. foo
	typemap map[string]string

	// ignoreSupport when set to true will prevent support templates from being
	// included in the generated source code. Support code is assumed to be provided
	// in another way.
	ignoreSupport bool
}

// NewPackage creates a package given a single directory where the source of
// the package lives.
func NewPackage(dir string) *Package {
	return &Package{
		Dir: dir,
		Config: &loader.Config{
			TypeChecker: types.Config{Error: func(err error) {}},
			TypeCheckFuncBodies: func(path string) bool {
				// only check function bodies in our own directory.
				if path == dir {
					return true
				}
				return false
			},
			ParserMode:  parser.ParseComments,
			AllowErrors: true,
		},
		generated: make(map[string]string),
		fileset:   make(map[string]*ast.File),
		filename:  template.Must(template.New("filename").Parse("Jig{{.Package}}{{.Name}}.go")),
		typemap:   make(map[string]string),
	}
}

// PkgSpec returns a PkgSpec instance with Path and Files correctly initialized.
func (p *Package) PkgSpec() loader.PkgSpec {
	var files []*ast.File
	for _, file := range p.fileset {
		files = append(files, file)
	}
	return loader.PkgSpec{Path: p.Dir, Files: files}
}

func (p *Package) Typemap() map[string]string {
	return p.typemap
}

// Filepath will return the filepath for a given *ast.File param.
func (p *Package) Filepath(file *ast.File) string {
	return p.Fset.File(file.Package).Name()
}

// AddFile is used by ParseDir() and GenerateSource(). It replaces the file
// stored by path in the fileset. The call is idempotent.
func (p *Package) AddFile(file *ast.File) {
	path := p.Filepath(file)
	p.fileset[path] = file
}

// WriteFileset writes the set of files that contains generated source to disk.
func (p *Package) WriteFileset(fileset map[*ast.File]struct{}) error {
	var err error
	for file := range fileset {
		path := p.Filepath(file)

		// create a file to write the new package content to.
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// actually write the package content out to file.
		if e := p.WriteFile(f, file); e != nil {
			err = e
		}
	}
	return err
}

// WriteFile will write the given file to the given output.
func (p *Package) WriteFile(output io.Writer, file *ast.File) error {
	return printer.Fprint(output, p.Fset, file)
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
