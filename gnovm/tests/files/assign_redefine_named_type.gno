package main

// `:=` that names an existing variable assigns to it instead of redeclaring
// it: the variable keeps its type, and the value converts as `a = x` would.
// See https://github.com/gnolang/gno/issues/6493.

type NamedMap map[string]int

func twoMaps() (map[string]int, map[string]int) {
	return map[string]int{"a": 1}, map[string]int{"b": 2}
}

func oneMap() map[string]int {
	return map[string]int{"a": 1}
}

func main() {
	var a NamedMap
	a, b := twoMaps()
	println(len(a), len(b))

	var c NamedMap
	c, d := oneMap(), 2
	println(len(c), d)
}

// Output:
// 1 1
// 1 2
