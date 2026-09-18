// Pins the pointer-equality contract for zero-sized types documented in
// docs/resources/go-gno-compatibility.md § Pointer equality for zero-sized
// types. Gno is uniform (Base, Index) identity for every element type — no
// runtime.zerobase folding, no offset-arithmetic collapse — so rows that
// gc-Go may report as true are always false here.
package main

type T struct{}

func main() {
	// Divergence rows: false in Gno; gc-Go varies (escape-dependent for
	// new/var, unconditionally true for the array/struct offset cases).
	println(new(T) == new(T))
	var x, y T
	println(&x == &y)
	var a [10]T
	println(&a[0] == &a[1])
	var s struct{ a, b T }
	println(&s.a == &s.b)

	// Same-identity rows: true in both Gno and gc-Go.
	println(&x == &x)
	println(&a[3] == &a[3])
	p := &x
	q := p
	println(p == q)
}

// Output:
// false
// false
// false
// false
// true
// true
// true
