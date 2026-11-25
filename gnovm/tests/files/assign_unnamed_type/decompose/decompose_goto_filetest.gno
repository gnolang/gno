package main

// A forward goto over the converting assignment lands on its label.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

func main() {
	var m map[string]int
	var ok bool
	i := 0
	goto L
	m, ok = fetch()
	println("skipped")
L:
	println("at L", i, ok, len(m))
}

// Output:
// at L 0 false 0
