package pkg

import (
	"go/ast"
	"go/token"
	"reactivego/jig/templ"
	"regexp"
	"strings"
)

// supportsAll is the constant used in supports to indicate it supports all other jigs.
const supportsAll = ""

// Extract template variables, matches will contain [["<Foo>" "Foo"] ["<Bar>" "Bar"]]
var reTemplateVar = regexp.MustCompile("<([[:word:]]+)>")

// jig contains everything from templ.Template and transient stuff not needed after analyzing all ast files.
type jig struct {
	templ.Template
	Pos    token.Pos
	End    token.Pos
	Source string

	// supports contains the templates this one supports, can be converted
	// into Needs on other templates. If it contains a single empty string
	// then this templates is needed by all other templates.
	// example: Observable<Foo> Concat, Observable<Foo> Merge
	supports []string
}

func newJig(packageName string, cgroup *ast.CommentGroup) *jig {
	jig := &jig{}
	jig.PackageName = packageName
	jig.Pos = cgroup.End()
	for _, comment := range cgroup.List {
		kvmatch := reCommentPragma.FindStringSubmatch(comment.Text)
		if len(kvmatch) == 3 {
			switch kvmatch[1] {
			case jigTemplate:
				jig.Name = kvmatch[2]
			case jigNeeds:
				needs := strings.Split(kvmatch[2], ",")
				for _, need := range needs {
					jig.Needs = append(jig.Needs, strings.TrimSpace(need))
				}
			case jigEmbeds:
				embeds := strings.Split(kvmatch[2], ",")
				for _, embed := range embeds {
					jig.Embeds = append(jig.Embeds, strings.TrimSpace(embed))
				}
			case jigSupports:
				supports := strings.Split(kvmatch[2], ",")
				for _, support := range supports {
					support = strings.TrimSpace(support)
					// Handle supports * case.
					if support == "*" {
						jig.setSupports(supportsAll)
						break
					}
					// Handle normal supports case.
					jig.supports = append(jig.supports, support)
				}
			}
		}
	}

	// Extract template vars e.g. [["<Foo>" "Foo"] ["<Bar>" "Bar"]]
	tplvars := reTemplateVar.FindAllStringSubmatch(jig.Name, -1)

	// Assign the template vars used in this template.
	// e.g. for tplvars == [["<Foo>" "Foo"] ["<Bar>" "Bar"]] assign Vars = ["Foo","Bar"]
	for _, tplvar := range tplvars {
		jig.Vars = append(jig.Vars, tplvar[1])
	}

	return jig
}

func (jig *jig) Close(pos token.Pos) {
	if jig != nil {
		jig.End = pos
	}
}

func (jig *jig) Supports(value string) bool {
	return len(jig.supports) == 1 && jig.supports[0] == ""
}

func (jig *jig) setSupports(value string) {
	jig.supports = []string{value}
}

func (jig *jig) ContainsSourceRange(pos, end token.Pos) bool {
	return pos > jig.Pos && end < jig.End
}

func (jig *jig) AddSource(source string) {
	if jig.Source != "" {
		jig.Source += "\n"
	}
	jig.Source += source
	jig.Source += "\n"
}
