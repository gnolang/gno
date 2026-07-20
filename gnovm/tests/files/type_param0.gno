package main

type W[P any] struct {
	a, b [0]P
}

func main() {
	println("unreachable")
}

// Gno targets go1.17 and does not support generics: the go/types pass rejects
// the type-parameter declaration before it runs (TypeCheckError below), which is
// what stops a generic type-expansion fan-out from ever reaching validType.

// Error:
// main/type_param0.gno:3:6-5:2: name P not defined in fileset with files [type_param0.gno]

// TypeCheckError:
// main/type_param0.gno:3:6: generic type declarations are not supported (Gno targets go1.17)
