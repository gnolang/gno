package main

// Alias chains that close on a defined type through a slice, and an alias
// of a defined type referenced from that type's own declaration.

type (
	a = b
	b = c
	c []a
)

type S struct{ p *T }

type T = S

func main() {
	var x c
	var s S
	println(len(x), s.p == nil)
}

// Output:
// 0 true
