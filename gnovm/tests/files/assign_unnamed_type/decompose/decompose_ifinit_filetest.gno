package main

// The converting assignment is allowed in an if-init.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

func main() {
	var m map[string]int
	var ok bool
	if m, ok = fetch(); ok {
		println("if", m["a"])
	}
	println(len(m), ok)
}

// Output:
// if 1
// 1 true
