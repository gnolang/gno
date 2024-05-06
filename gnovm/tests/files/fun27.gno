package main

type Foo interface {
	foo()
}

func NewSet() Foo {
	return 1
}

func main() {
	NewSet()
}

// Error:
// main/files/fun27.gno:9: int does not implement test.Foo
