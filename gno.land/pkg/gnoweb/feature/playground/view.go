package playground

// PlaygroundData is the render payload for templates/page.html.
type PlaygroundData struct {
	// InitialCode is a shared snippet (?code=), in the "// --- name ---"
	// text format the Share button writes.
	InitialCode string

	// Files is a package copy (?from=), one entry per file: never put
	// through the text format, which a source line could split.
	Files []File

	// From is the copied package's path without its domain (r/demo/foo),
	// empty for anything but a copy.
	From string

	// Domain is the chain domain.
	Domain string

	// DefaultFile is the filename that should be focused on first load.
	DefaultFile string

	// Preview is true when Run shows what the copied realm renders,
	// instead of the commands that run the code locally.
	Preview bool
}
