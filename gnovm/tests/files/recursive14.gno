package main

// A defined type over a name that is still being built. B needs only A's
// shell, but A copies B's base, so B must be built first whichever
// declaration is predefined first.

type B struct{ p *A }

type A B

type C D

type D struct{ c *C }

func main() {
	var a A
	var c C
	println(a.p == nil, c.c == nil)
}

// Output:
// true true
