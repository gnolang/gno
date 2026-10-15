package omnisearch

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// funcMap holds only what these templates need. It deliberately does not
// import the components package's map: that one is package-private there, and
// a shared map would couple the feature's rendering to the core's.
var funcMap = template.FuncMap{
	"isIndexer": func(s Source) bool { return s == SourceIndexer },
	"errText": func(err error) string {
		if err == nil {
			return ""
		}
		return err.Error()
	},
}

// pageTemplate is parsed once at init: a malformed template surfaces
// immediately rather than on the first request.
var pageTemplate = template.Must(
	template.New("renderSearch").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"),
)
