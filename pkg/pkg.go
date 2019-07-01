package pkg

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/types"
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

	// forceCommon (default set to false) forces common code templates (i.e. not
	// specialized on type) to be included in the generated source code. Common
	// code normally assumed to be present already in the package providing the
	// templates. Common code templates are templates either explicitly marked
	// as such (jig:common) and templates that are marked as needed by another
	// template but that don't have template vars themselves.
	forceCommon bool
}

// NewPackage creates a package given a single directory where the source of
// the package lives.
func NewPackage(dir string) *Package {
	buildConfig := build.Default
	buildConfig.CgoEnabled = false

	conf := &loader.Config{
		TypeChecker: types.Config{Error: func(err error) {  }},
		TypeCheckFuncBodies: func(path string) bool {
			// only check function bodies in our own directory.
			if path == dir {
				return true
			}
			return false
		},
		ParserMode:  parser.ParseComments,
		AllowErrors: true,
		Build:       &buildConfig,
	}
	return &Package{
		Dir: dir,
		Config: conf,
		generated: make(map[string]string),
		fileset:   make(map[string]*ast.File),
		filename:  template.Must(template.New("filename").Parse("{{.package}}.go")),
		typemap:   make(map[string]string),
	}
}

// PkgSpec returns PkgSpec instances with Path and Files correctly initialized.
func (p *Package) PkgSpec() []loader.PkgSpec {
	var files []*ast.File
	for _, file := range p.fileset {
		if p.Name == file.Name.String() {
			files = append(files, file)
		}
	}
	return []loader.PkgSpec{ loader.PkgSpec{Path: p.Dir, Files: files} }
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

// GeneratedFileset returns a map containing the set of files
// that have generated fragments in them.
func (p *Package) GeneratedFileset() map[*ast.File]struct{} {
	fileset := make(map[*ast.File]struct{})
	for _, path := range p.generated {
		file := p.fileset[path]
		fileset[file] = struct{}{}
	}
	return fileset
}
