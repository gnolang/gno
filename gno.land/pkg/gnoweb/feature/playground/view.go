package playground

// PlaygroundData is the render payload for templates/page.html.
type PlaygroundData struct {
	// InitialCode is pre-filled code (a package copy from ?from= or a shared snippet).
	InitialCode string

	// Domain is the chain domain.
	Domain string

	// DefaultFile is the filename that should be focused on first load.
	DefaultFile string
}
