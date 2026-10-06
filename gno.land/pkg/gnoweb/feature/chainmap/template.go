package chainmap

import (
	"embed"
	"html/template"
	"io"

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
).ParseFS(components.SharedPartialsFS(), "listing_truncated.html", "indexer_status.html", "pkg_graph.html"))

// pageComponent renders one of this feature's pages, so IndexLayout wraps it
// in the standard chrome without components knowing the template.
type pageComponent struct {
	name string
	data any
}

func (c *pageComponent) Render(w io.Writer) error {
	return pageTemplate.ExecuteTemplate(w, c.name, c.data)
}
