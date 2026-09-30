package playground

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// PageTemplate is the standalone playground page. Pre-parsed at init
// so a misconfigured template surfaces immediately, not on the first
// request.
var PageTemplate = mustParse("templates/page.html")

func mustParse(paths ...string) *template.Template {
	t, err := template.New("").ParseFS(templateFS, paths...)
	if err != nil {
		panic("playground: parse " + paths[0] + ": " + err.Error())
	}
	return t
}
