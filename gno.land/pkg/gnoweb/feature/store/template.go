package store

import (
	"embed"
	"html/template"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

//go:embed templates/*.html
var templateFS embed.FS

// newTemplates parses the store templates bound to the configured realm:
// storeURL "" is its front page, storeURL "c/defi" a render path, and
// storePkg its package path and domain the chain's, as builders import and
// call them.
func newTemplates(realmPath, domain string) *template.Template {
	return template.Must(template.New("store").Funcs(template.FuncMap{
		"storeURL": func(args string) string { return weburl.GnoURL{Path: realmPath, Args: args}.EncodeURL() },
		"storePkg": func() string { return domain + realmPath },
		"domain":   func() string { return domain },
		// snippet pairs a code block with the id its copy button targets.
		"snippet": func(id, code string) any { return struct{ ID, Code string }{id, code} },
		// lens names the current lens, and whether the page is the lens
		// itself rather than one under it.
		"lens": func(section string, exact bool) any {
			return struct {
				Section string
				Exact   bool
			}{section, exact}
		},
		// grid pairs cards with the grid's modifier class, if any.
		"grid": func(mod string, cards []card) any {
			return struct {
				Mod   string
				Cards []card
			}{mod, cards}
		},
	}).ParseFS(templateFS, "templates/*.html"))
}

// ViewType tags store views for the layout and analytics.
const ViewType components.ViewType = "store-view"

type pageComponent struct {
	tmpl *template.Template
	name string
	data any
}

func (c pageComponent) Render(w io.Writer) error {
	return c.tmpl.ExecuteTemplate(w, c.name, c.data)
}
