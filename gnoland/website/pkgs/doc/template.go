package doc

import (
	"fmt"
	"go/doc/comment"
)

func commentFmt(c string) string {
	var p comment.Parser
	pr := &comment.Printer{
		DocLinkBaseURL: "/p",
	}
	return string(pr.Markdown(p.Parse(c)))
}

func codeFmt(code string) string {
	return fmt.Sprintf("%s\n%s\n%s", "```go", code, "```")
}

var TemplateMarkdown = `
# Package {{ .Name }}

import "{{ .ImportPath }}"

## Overview

{{ if .Doc }}
{{ comment .Doc }}
{{ else }}
This section is empty.
{{ end }}

## Constants

{{ range .Consts }}
{{ if .Doc }}
{{ comment .Doc }}
{{ end }}
{{ code .Signature }}
{{ else }}
This section is empty.
{{ end }}

## Variables

{{ range .Vars }}
{{ if .Doc }}
{{ comment .Doc }}
{{ end }}
{{ code .Signature }}
{{ else }}
This section is empty.
{{ end }}

## Functions

{{ range .Funcs }}
### func {{ . }}
{{ code .Signature }}
{{ if .Doc }}
{{ comment .Doc }}
{{ end }}
{{ else }}
This section is empty.
{{ end }}

## Types

{{ range .Types }}
### <a name="{{ .ID }}"></a>type {{ .Name }}
{{ if .Doc }}
{{ comment .Doc }}
{{ end }}
{{ code .Definition }}

{{ range .Vars }}
<a name="{{ .ID }}"></a>
{{ if .Doc }}
{{ comment .Doc }}
{{ end }}
{{ code .Signature }}
{{ end }}

{{ range .Consts }}
<a name="{{ .ID }}"></a>
{{ if .Doc }}
{{ comment .Doc }}
{{ end }}
{{ code .Signature }}
{{ end }}

{{ range .Funcs }}
### <a name="{{ .ID }}"></a>func {{ . }} 
{{ code .Signature }}
{{ if .Doc }}
{{ .Doc }}
{{ end }}
{{ end }}

{{ range .Methods }}
### <a name="{{ .ID }}"></a>func {{ . }}
{{ code .Signature }}
{{ if .Doc }}
{{ comment .Doc }}
{{ end }}
{{ end }}
{{ else }}
This section is empty.
{{ end }}

## Source Files
{{ range .Filenames }}
- [{{ . }}]({{ $.Path }}/{{ . }})
{{ end }}
`
