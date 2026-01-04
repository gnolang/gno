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
// main/fun27.gno:8:2-10: <untyped> bigint does not implement main.Foo (missing method foo)

// TypeCheckError:
// main/fun27.gno:8:9: cannot use 1 (constant of type int) as Foo value in return statement: int does not implement Foo (missing method foo)
