package context

// XXX dummy
type Context interface {
	Err() error
}

// XXX dummy
type emptyCtx int

// XXX dummy
func (*emptyCtx) Err() error {
	return nil
}

// XXX dummy
var (
	background0 emptyCtx
	background  = &background0 // XXX replace with new(emptyCtx)
)

func Background() Context {
	return background
}
