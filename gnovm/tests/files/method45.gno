// Verifies that an ambiguity at the shallowest depth makes the selector
// illegal even when a unique match exists deeper (Go spec §Selectors:
// resolution is decided at the shallowest depth containing the name; a
// deeper match cannot rescue it).
package main

type M1 struct{ Foo string }
type M2 struct{ Foo string }

type AmbBox struct {
	M1
	M2
}

type Inner3 struct{ Foo string }
type Inner2 struct{ Inner3 }

type DeepBox struct {
	Inner2
}

type Outer struct {
	AmbBox
	DeepBox
}

func main() {
	var o Outer
	_ = o.Foo
}

// Error:
// main/method45.gno:29:6-11: ambiguous selector Foo in main.Outer

// TypeCheckError:
// main/method45.gno:29:8: ambiguous selector o.Foo
