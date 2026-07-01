// Verifies that selector resolution on an embedded field/method walks the
// embedding hierarchy BREADTH-FIRST by depth (Go spec §Selectors): the
// shallowest match wins, not the first match encountered in field-order
// depth-first search.
package main

type embedded string

func (s embedded) val() string {
	return string(s)
}

type A struct {
	embedded
}

type B struct {
	A
	embedded
}

func main() {
	b := &B{
		A: A{
			embedded: "a",
		},
		embedded: "b",
	}
	s := b.val()
	if s != "b" {
		panic(s)
	}
}

// Output:
