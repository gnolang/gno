package main

// A function type, like a pointer, refers to its parameter and result types
// through an indirection, so a cycle through a function type (or an
// interface method signature) and a defined type is valid Go. See
// golang/go#25838.

type (
	A = func() B
	B A
)

type (
	C = interface{ M() D }
	D C
)

type impl struct{}

func (impl) M() D { return impl{} }

func main() {
	var b B = func() B { return nil }
	println(b() == nil)

	var d D = impl{}
	println(d.M() != nil)
}

// Output:
// true
// true
