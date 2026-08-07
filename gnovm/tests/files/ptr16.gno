// https://github.com/gnolang/gno/issues/5957 (review follow-up)
// Methods are not selectable through a defined pointer type, but they
// still participate in depth/ambiguity resolution: a method and a
// field at the same shallowest depth make the selector ambiguous.
package main

type M struct{}

func (M) F() string { return "foo" }

type V struct{ F int }

type D2 struct {
	M
	V
}

type D1 *D2

func main() {
	d2 := D2{V: V{F: 7}}
	var x D1 = &d2
	println(x.F)
}

// Error:
// main/ptr16.gno:23:10-13: ambiguous selector F in main.D1

// TypeCheckError:
// main/ptr16.gno:23:12: ambiguous selector x.F
