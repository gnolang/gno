package main

// The converting assignment keeps its label as a goto target.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

func main() {
	var m map[string]int
	var ok bool
	i := 0
L:
	m, ok = fetch()
	i++
	if i < 2 {
		goto L
	}
	println(i, m["a"], ok)
}

// Output:
// 2 1 true
