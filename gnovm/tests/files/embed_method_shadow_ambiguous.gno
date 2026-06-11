package main

// Two methods of the same name promoted from sibling embedded types at the
// same (shallowest) depth make the selector ambiguous, so the struct does
// not implement the interface (Go spec, "Selectors").

type I interface{ M() string }

type A struct{}

func (A) M() string { return "a" }

type B struct{}

func (B) M() string { return "b" }

type T struct {
	A
	B
}

func main() {
	var i I = T{}
	println(i.M())
}

// Error:
// main/embed_method_shadow_ambiguous.gno:23:6-15: main.T does not implement main.I (missing method M)

// TypeCheckError:
// main/embed_method_shadow_ambiguous.gno:23:12: cannot use T{} (value of struct type T) as I value in variable declaration: T does not implement I (ambiguous selector T.M)
