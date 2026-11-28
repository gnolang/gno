package main

// The type-assertion form of comma-ok has the same untyped bool result as the
// map-index form, so its ok is assignable to a declared bool.
// See https://github.com/gnolang/gno/issues/6493.

type B bool

func main() {
	var x any = B(true)
	var v, ok B = x.(B)
	println(v, ok)
}

// Output:
// (true main.B) (true main.B)
