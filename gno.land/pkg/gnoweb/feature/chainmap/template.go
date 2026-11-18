package chainmap

import (
	"embed"
	"html/template"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

//go:embed templates/*.html
var templateFS embed.FS

// pageTemplate is parsed once at init: a malformed template surfaces
// immediately rather than on the first request. The truncation notice, the
// provenance footer and the dependency graph are shared partials, so these
// pages say each exactly as the list and the overview do.
var pageTemplate = template.Must(template.Must(
	template.New("chainmap").ParseFS(templateFS, "templates/*.html"),
).ParseFS(components.SharedPartialsFS(), "indexer_status.html", "pkg_graph.html"))

// page renders the template name of this feature's set, so IndexLayout
// wraps it in the standard chrome without components knowing the template.
func page(name string, data any) components.Component {
	return &components.FeatureComponent{Tmpl: pageTemplate, Name: name, Data: data}
}
