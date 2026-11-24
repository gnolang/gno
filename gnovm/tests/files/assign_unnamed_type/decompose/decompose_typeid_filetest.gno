package main

// Each result takes its Lhs type, named or unnamed, in a body, a loop and a
// closure.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

func show(x any) {
	switch x.(type) {
	case NamedMap:
		println("NamedMap")
	case Named2:
		println("Named2")
	case map[string]int:
		println("map")
	default:
		println("other")
	}
}

type Named2 map[string]int

func fetch2() (map[string]int, bool) { return map[string]int{"b": 2}, true }

func main() {
	var m map[string]int
	var ok bool
	m, ok = fetch()
	show(m)
	var n Named2
	n, ok = fetch2()
	show(n)
	var nn any = n
	_, isN2 := nn.(Named2)
	println(isN2, ok)
	for i := 0; i < 2; i++ {
		var m2 map[string]int
		m2, ok = fetch()
		show(m2)
	}
	f := func() {
		var m3 map[string]int
		m3, ok = fetch()
		show(m3)
	}
	f()
}

// Output:
// map
// Named2
// true true
// map
// map
// map
