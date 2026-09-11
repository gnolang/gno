// https://github.com/gnolang/gno/issues/5957 (review follow-up)
// A method at a shallower depth shadows a deeper field even through a
// defined pointer type: the winner is the method, which is then not
// selectable — the selector does not fall through to the deeper field.
package main

type M struct{}

func (M) F() string { return "foo" }

type Inner struct{ F int }

type W struct{ Inner }

type D2 struct {
	M
	W
}

type D1 *D2

func main() {
	var x D1
	_ = x.F
}

// Error:
// main/ptr17.gno:24:6-9: missing field F in main.D1

// TypeCheckError:
// main/ptr17.gno:24:8: x.F undefined (type D1 has no field or method F)
