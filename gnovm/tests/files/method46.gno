// Verifies that reaching the SAME type through two same-depth embedded
// paths (diamond embedding) makes its names ambiguous (Go spec
// §Selectors: not exactly one Foo at the shallowest depth), even though
// both paths agree on the type providing the name.
package main

type Inner struct{ Foo string }

type A struct{ Inner }
type B struct{ Inner }

type C struct {
	A
	B
}

func main() {
	var c C
	_ = c.Foo
}

// Error:
// main/method46.gno:19:6-11: ambiguous selector Foo in main.C

// TypeCheckError:
// main/method46.gno:19:8: ambiguous selector c.Foo
