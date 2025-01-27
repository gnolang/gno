// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

import "runtime"

var m = map[string]int{"a": 1}

type S struct {
	name string
	mm   map[string]int
}

func main() {
	var s = S{name: "foo"}
	s.mm = m
	runtime.GC()
}
