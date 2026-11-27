package main

// The converting assignment is allowed in a switch-init.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

func main() {
	var m map[string]int
	var ok bool
	switch m, ok = fetch(); ok {
	case true:
		println("sw", m["a"])
	}
	println(len(m), ok)
}

// Output:
// sw 1
// 1 true
