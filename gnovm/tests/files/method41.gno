// Verifies that two embedded siblings exposing the same field name at the
// same depth produce an ambiguous-selector error (Go spec §Selectors:
// "If there is not exactly one f with shallowest depth, the selector
// expression is illegal").
package main

type A struct{ Foo int }
type B struct{ Foo int }
type C struct {
	A
	B
}

func main() {
	var c C
	_ = c.Foo
}

// Error:
// main/method41.gno:16:6-11: ambiguous selector Foo in main.C

// TypeCheckError:
// main/method41.gno:16:8: ambiguous selector c.Foo
