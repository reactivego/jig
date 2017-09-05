package pkg

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"reactivego/jig/templ"
	"strings"
)

// LoadTemplates is used to go through all the ast.File(s) in all the
// packages and then turn all jigs that are found in those files into templates.
func (p *Package) LoadTemplates(tplr templ.Templater) (messages []string, err error) {
	for _, pkgInfo := range p.allPackages {
		var jigs []*jig
		for _, file := range pkgInfo.Files {
			jigs = append(jigs, p.LoadTemplatesFromFile(file)...)
		}
		if jigs == nil {
			continue
		}
		// Convert support declarations into Needs on the templates
		p.transformSupportIntoNeeds(jigs)
		// Now all jigs have been read, so we can now tell every jig to define their template.
		err = p.defineTemplates(tplr, jigs)
		if err != nil {
			return messages, err
		}
		messages = append(messages, fmt.Sprintf("found %d jig templates in package %q", len(jigs), pkgInfo.Pkg.Path()))
	}
	tplr.Sort()
	return messages, nil
}

// LoadTemplatesFromFile will parse comments in a file to find //jig:template entries that
// declare templates and use that to determine source range of the associated template definition.
// Then walk the file ast and extract source in the ranges determined before and add it to the
// correct jig.
func (p *Package) LoadTemplatesFromFile(file *ast.File) []*jig {
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
				jig = newJig(file.Name.String(), cgroup)
				jigs = append(jigs, jig)
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
	ast.Walk(sourceCollector{Fset: p.Fset, Snippets: jigs}, file)
}

// sourceCollector is used to visit all the nodes in an ast tree and collect
// and glue together source fragments that belong to jigs. After the visit is
// done all jigs have their complete source attached.
type sourceCollector struct {
	Fset     *token.FileSet
	Snippets []*jig
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
			if jig.ContainsSourceRange(decl.Pos(), decl.End()) {
				var source bytes.Buffer
				printer.Fprint(&source, c.Fset, decl)
				jig.AddSource(source.String())
			}
		}
	}
	return nil
}

// transformSupportIntoNeeds will convert support declarations of a jig into Needs on the other jigs
func (p *Package) transformSupportIntoNeeds(jigs []*jig) {
	if p.ignoreSupport {
		return
	}
	var supports []string
	for _, jig := range jigs {
		if jig.support {
			supports = append(supports, jig.Name)
		}
	}
	for _, jig := range jigs {
		if jig.support {
			continue
		}
		jig.Needs = append(supports, jig.Needs...)
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
