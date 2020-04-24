package pkg

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"

	"github.com/reactivego/generics/templ"
)

// LoadTemplates is used to go through all the ast.File(s) in all the
// packages and then turn all jigs that are found in those files into templates.
func (p *Package) LoadTemplates(tplr templ.Templater) (messages []string, err error) {
	for _, pkgInfo := range p.allPackages {
		ignoreSupportTemplates := !p.forceCommon && p.Dir == pkgInfo.Pkg.Path()
		var jigs []*jig
		for _, file := range pkgInfo.Files {
			jigs = append(jigs, p.LoadTemplatesFromFile(file, ignoreSupportTemplates)...)
		}
		if jigs == nil {
			continue
		}
		if !ignoreSupportTemplates {
			p.transformCommonIntoNeeds(jigs)
		}
		// Now all jigs have been read, so we can now tell every jig to define their template.
		err = p.defineTemplates(tplr, jigs)
		if err != nil {
			return messages, err
		}
		var msg string
		if !ignoreSupportTemplates {
			msg = fmt.Sprintf("found %d templates in package %q (%s)", len(jigs), pkgInfo.Pkg.Name(), pkgInfo.Pkg.Path())
		} else {
			msg = fmt.Sprintf("found %d templates in package %q (%s) ignoring support templates", len(jigs), pkgInfo.Pkg.Name(), pkgInfo.Pkg.Path())
		}
		messages = append(messages, msg)
	}
	tplr.Sort()
	return messages, nil
}

// LoadTemplatesFromFile will parse comments in a file to find //jig:template entries that
// declare templates and use that to determine source range of the associated template definition.
// Then walk the file ast and extract source in the ranges determined before and add it to the
// correct jig.
func (p *Package) LoadTemplatesFromFile(file *ast.File, ignoreSupportTemplates bool) []*jig {
	var jigs []*jig
	var jig *jig
	for _, cgroup := range file.Comments {
		for _, comment := range cgroup.List {
			// jig:end
			if strings.HasPrefix(comment.Text, jigEnd) {
				jig.Close(cgroup.Pos())
				jig = nil
			}
			// jig:template <name>
			if strings.HasPrefix(comment.Text, jigTemplate) {
				jig.Close(cgroup.Pos())
				packageName := file.Name.String()
				jig = newJig(packageName, cgroup)
				jigHasSupportingRole := jig.common || len(jig.Vars) == 0
				if !ignoreSupportTemplates || !jigHasSupportingRole {
					jigs = append(jigs, jig)
				}
				break
			}
		}
	}
	jig.Close(file.End() + 1)
	// Collect the source for the jigs.
	p.collectSources(jigs, file)

	return jigs
}

// collectSources will visit all declarations in the file and collect the source for the jigs.
func (p *Package) collectSources(jigs []*jig, file *ast.File) {
	ast.Walk(sourceCollector{Fset: p.Fset, Snippets: jigs, Nodoc: p.nodoc}, file)
}

// sourceCollector is used to visit all the nodes in an ast tree and collect
// and glue together source fragments that belong to jigs. After the visit is
// done all jigs have their complete source attached.
type sourceCollector struct {
	Fset     *token.FileSet
	Snippets []*jig
	Nodoc bool
}

// Visit a specific ast node and add the source representation of
// that declaration to the jig for which the Pos and End postion
// encapsulates the declaration's Pos and End position.
func (c sourceCollector) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch decl := node.(type) {
	case *ast.File:
		return c

	case *ast.GenDecl, *ast.FuncDecl:
		for _, jig := range c.Snippets {
			pos := decl.Pos()
			// skip decl doc
			if c.Nodoc {
				if g,ok := node.(*ast.GenDecl); ok {
					g.Doc = nil
				}
				if f,ok := node.(*ast.FuncDecl); ok {
					f.Doc = nil
				}
			}
			if jig.ContainsSourceRange(pos, decl.End()) {
				var source bytes.Buffer
				printer.Fprint(&source, c.Fset, decl)
				jig.AddSource(source.String())
			}
		}
	}
	return nil
}

// transformCommonIntoNeeds will convert common declarations of a jig
// into Needs on the other jigs
func (p *Package) transformCommonIntoNeeds(jigs []*jig) {
	var common []string
	for _, jig := range jigs {
		if jig.common {
			common = append(common, jig.Name)
		}
	}
	for _, jig := range jigs {
		if jig.common {
			continue
		}
		jig.Needs = append(common, jig.Needs...)
	}
}

// defineTemplates will add a template for every jig.
func (p *Package) defineTemplates(tplr templ.Templater, jigs []*jig) error {
	for _, jig := range jigs {
		err := tplr.Add(jig.Template, jig.Source)
		if err != nil {
			return err
		}
	}
	return nil
}
