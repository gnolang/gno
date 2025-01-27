// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

import "runtime"

type S struct {
	name string
	mm   map[string]int
}

// XXX, the rhs structLit is allocated first,
// with struct fields and struct itself
// then, with the assign, the struct if
// copied, then it's allocated again.
var s = S{name: "foo"}

func main() {
	s.mm = map[string]int{"a": 1}
	runtime.GC()
}
