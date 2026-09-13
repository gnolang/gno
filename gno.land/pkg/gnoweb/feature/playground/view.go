package playground

// PlaygroundData is the render payload for templates/page.html.
type PlaygroundData struct {
	// InitialCode is pre-filled code (e.g. from fork or shared snippet).
	InitialCode string

	// ForkFrom is the package path this was forked from, when the page
	// is rendered as a fork. Empty for a fresh playground.
	ForkFrom string

	// Remote is the RPC endpoint.
	Remote string

	// ChainId is the current chain ID.
	ChainId string

	// Domain is the chain domain.
	Domain string

	// DefaultFile is the filename that should be focused on first load.
	DefaultFile string
}
