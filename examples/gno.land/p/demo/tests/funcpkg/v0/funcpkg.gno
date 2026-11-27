package funcpkg

// Top is a top-level function declared in this library.
func Top() {}

// BuildClosure returns a closure built by this library.
func BuildClosure() func() {
	return func() {}
}

// T is a library type with a method.
type T struct{}

// M is a method declared in this library.
func (T) M() {}

// NewT returns a T.
func NewT() T {
	return T{}
}
