package main

// A field promoted from one embedded type and a method of the same name
// promoted from another at the same depth make the selector ambiguous.

type A struct{ X string }

type B struct{}

func (B) X() string { return "b" }

type T struct {
	A
	B
}

func main() {
	t := T{}
	_ = t.X
}

// Error:
// main/embed_method_shadow_fieldmethod.gno:19:6-9: missing field X in main.T

// TypeCheckError:
// main/embed_method_shadow_fieldmethod.gno:19:8: ambiguous selector t.X
