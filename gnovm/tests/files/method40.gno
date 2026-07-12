// Verifies that selector resolution on an embedded field/method walks the
// embedding hierarchy BREADTH-FIRST by depth (Go spec §Selectors): the
// shallowest match wins, not the first match encountered in field-order
// depth-first search.
//
// Program taken verbatim from go/test/fixedbugs/bug485.go, which pinned the
// same bug in gccgo (2014): "Gccgo chose the wrong embedded method when the
// same type appeared at different levels and the correct choice was not the
// first appearance of the type in a depth-first search." GnoVM independently
// reproduced it: b.val() resolved to the deeper B.A.embedded ("a") instead of
// the shallower B.embedded ("b"), silently dispatching the wrong receiver.
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
