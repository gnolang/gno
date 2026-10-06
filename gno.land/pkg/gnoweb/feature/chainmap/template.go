package chainmap

import (
	"embed"
	"html/template"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

//go:embed templates/*.html
var templateFS embed.FS

// mapTemplate is parsed once at init: a malformed template surfaces
// immediately rather than on the first request. The truncation notice and the
// provenance footer are the shared partials, so the map says both exactly as
// the list and the search page do.
var mapTemplate = template.Must(template.Must(
	template.New("renderMap").ParseFS(templateFS, "templates/*.html"),
).ParseFS(components.SharedPartialsFS(), "listing_truncated.html", "indexer_status.html"))
