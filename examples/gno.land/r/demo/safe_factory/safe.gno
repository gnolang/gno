package safe

// XXX: consider splitting into multiples files or packages

// FooSafe is a safe object... meaning...
type FooSafe struct {
	name string
	counter int
}

func (f *FooSafe) Inc() { f.counter++ }

func NewSafe(name string) *FooSafe {
	return &FooSafe{name: name}
}

// XXX: external realm that calls the factory and store the pointer
// XXX: demo with maketx run registering on a directory
