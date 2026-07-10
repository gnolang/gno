// Gno targets go1.17 and rejects go1.18 generics syntax at deploy-time
// type-checking; see gnovm/pkg/gnolang/nogenerics.go (checkNoGenerics).

package main

type Box[T any] struct{ v T }

func main() {
	println("unreached")
}

// Error:
// main/type_param0.gno:6:6-30: name T not defined in fileset with files [type_param0.gno]

// TypeCheckError:
// main/type_param0.gno:6:6: generic type declarations are not supported (Gno targets go1.17)
