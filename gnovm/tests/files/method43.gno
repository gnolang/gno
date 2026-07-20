// Verifies that a promoted METHOD and a promoted FIELD with the same
// name at the same depth are ambiguous (Go spec §Selectors makes no
// distinction between fields and methods for depth resolution).
package main

type A struct{}

func (A) Foo() string { return "method" }

type B struct{ Foo string }

type C struct {
	A
	B
}

func main() {
	var c C
	_ = c.Foo
}

// Error:
// main/method43.gno:19:6-11: ambiguous selector Foo in main.C

// TypeCheckError:
// main/method43.gno:19:8: ambiguous selector c.Foo
