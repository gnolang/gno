package run

import (
	"embed"
	"html/template"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

//go:embed templates/*.html
var templateFS embed.FS

// PageTemplate is the run scratchpad page. Pre-parsed at init so a
// misconfigured template surfaces immediately, not on the first request.
var PageTemplate = mustParse("templates/page.html")

func mustParse(paths ...string) *template.Template {
	t, err := template.New("").ParseFS(templateFS, paths...)
	if err != nil {
		panic("run: parse " + paths[0] + ": " + err.Error())
	}

	// Reuse component UI partials
	if t, err = t.ParseFS(components.TemplatesFS(), "ui/btn_copy.html"); err != nil {
		panic("run: parse shared partials: " + err.Error())
	}
	return t
}
