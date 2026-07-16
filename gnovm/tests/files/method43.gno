// https://github.com/gnolang/gno/issues/5957
// Fields (but not methods) promote through a defined pointer type;
// methods remain reachable through an explicit deref or field selector.
package main

type E struct{}

func (E) Bar() string { return "bar" }

type D2 struct {
	E
	A int
}

func (D2) Foo() string { return "m" }

type D1 *D2

func main() {
	d2 := D2{A: 5}
	var x D1 = &d2
	println(x.A)          // field: promoted via (*x).A
	println(x.E.Bar())    // method on explicitly selected field
	println((*x).Foo())   // method via explicit deref
}

// Output:
// 5
// bar
// m
