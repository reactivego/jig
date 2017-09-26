package pkg

import (
	"fmt"
	"go/ast"
	"strings"
	"text/template"
)

func (p *Package) LoadGeneratePragmas() (messages []string) {
	pkgInfo := p.PkgSpec()
	for _, file := range pkgInfo.Files {
		msgs := p.loadGeneratePragmasFromFile(file)
		messages = append(messages, msgs...)
	}
	return messages
}

func (p *Package) loadGeneratePragmasFromFile(file *ast.File) (messages []string) {
	for _, cgroup := range file.Comments {
		for _, comment := range cgroup.List {
			// jig:file <filename>
			if strings.HasPrefix(comment.Text, jigFile) {
				kvmatch := reCommentPragma.FindStringSubmatch(comment.Text)
				if len(kvmatch) == 3 && kvmatch[2] != "" {
					filename, err := template.New("filename").Parse(kvmatch[2])
					if err == nil {
						p.filename = filename
					} else {
						messages = append(messages, fmt.Sprintf("ignoring pragma //jig:file %s", kvmatch[2]), err.Error())
					}
				}
			}
			// jig:type <display type> <real type>
			if strings.HasPrefix(comment.Text, jigType) {
				kvmatch := reJigType.FindStringSubmatch(comment.Text)
				if len(kvmatch) == 3 && kvmatch[1] != "" && kvmatch[2] != "" {
					p.typemap[kvmatch[1]] = kvmatch[2]
				}
			}
			// jig:no-support
			if strings.HasPrefix(comment.Text, jigNoSupport) {
				p.ignoreSupport = true
			}
		}
	}
	return messages
}
