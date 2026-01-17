package lint

type Reporter interface {
	Report(issue Issue)
	Flush() error
	Summary() (info, warnings, errors int)
}
