package playground

// PlaygroundData is the render payload for templates/page.html.
type PlaygroundData struct {
	// InitialCode is a shared snippet (?code=), in the "// --- name ---"
	// text format the Share button writes.
	InitialCode string

	// Files is a package copy (?from=), one entry per file: never put
	// through the text format, which a source line could split.
	Files []File

	// Domain is the chain domain.
	Domain string

	// DefaultFile is the filename that should be focused on first load.
	DefaultFile string
}
