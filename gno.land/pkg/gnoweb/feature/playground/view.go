package playground

// PlaygroundData is the render payload for templates/page.html.
type PlaygroundData struct {
	// Files is the code the editor opens, one <pre> per file: a package
	// copy (?from=), a shared snippet (?code=) or the default code.
	Files []File

	// From is the copied package's path without its domain (r/demo/foo),
	// empty for anything but a copy.
	From string

	// Domain is the chain domain.
	Domain string

	// DefaultFile is the filename that should be focused on first load.
	DefaultFile string

	// Refused says why the ?code= link was not opened, when it was not:
	// the page shows the default code, and says so.
	Refused string

	// Preview is true when Run shows what the copied realm renders,
	// instead of the commands that run the code locally.
	Preview bool
}
