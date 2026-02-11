package lint

// XXX: Consider moving Reporter to a standalone package (e.g., gnovm/pkg/report)
// to allow reuse by other commands (run, test) without coupling them to the lint package.
type Reporter interface {
	Report(issue Issue)
	Flush() error
	Summary() (info, warnings, errors int)
}
