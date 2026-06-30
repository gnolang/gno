package main

// If an embedded type is itself internally ambiguous for the method at the
// shallowest matching depth, the whole selector is ambiguous even when a
// sibling embedded type has a clean match at that same depth. Here Amb's M
// (from A and B) and Clean's M (from A) are all at the same depth, so T
// does not implement I.

type I interface{ M() string }

type A struct{}

func (A) M() string { return "A" }

type B struct{}

func (B) M() string { return "B" }

type Amb struct {
	A
	B
}

type Clean struct{ A }

type T struct {
	Amb
	Clean
}

func main() {
	var i I = T{}
	println(i.M())
}

// Error:
// main/embed_method_shadow_internal_ambiguous.gno:32:6-15: main.T does not implement main.I (missing method M)

// TypeCheckError:
// main/embed_method_shadow_internal_ambiguous.gno:32:12: cannot use T{} (value of struct type T) as I value in variable declaration: T does not implement I (ambiguous selector T.M)
