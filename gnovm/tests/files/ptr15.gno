// https://github.com/gnolang/gno/issues/5957 (review follow-up)
// A defined pointer type has an empty method set, so it does not
// satisfy an interface through its base's methods.
package main

type D2 struct{}

func (D2) Foo() string { return "m" }

type D1 *D2 // empty method set: does not implement I

type I interface{ Foo() string }

func main() {
	var x D1
	var _ I = x
	println("unreached")
}

// Error:
// main/ptr15.gno:16:6-13: main.D1 does not implement main.I (missing method Foo)

// TypeCheckError:
// main/ptr15.gno:16:12: cannot use x (variable of pointer type D1) as I value in variable declaration: D1 does not implement I (missing method Foo)
