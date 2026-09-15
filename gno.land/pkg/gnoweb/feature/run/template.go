package run

import (
	"embed"
	"io/fs"
)

//go:embed templates/*.html
var templateFS embed.FS

// TemplatesFS exposes the run scratchpad partial so the components package can
// parse it into the shared "web" template set. The parse happens once in that
// package's init (components/template.go), which panics on a malformed
// template, so errors here surface at startup rather than on first request.
//
// The dependency runs components -> run, the reverse of the other features:
// this package holds templates only, so it must not import components (that
// would cycle). Keep it free of Go dependencies on the rest of gnoweb.
func TemplatesFS() fs.FS { return templateFS }
