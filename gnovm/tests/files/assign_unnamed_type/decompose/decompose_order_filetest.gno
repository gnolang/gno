package main

// Go evaluates the Lhs index operands before the call on the right.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { println("fetch"); return NamedMap{"a": 1}, true }

func idx() int { println("idx"); return 0 }

func main() {
	var arr [1]map[string]int
	var ok bool
	arr[idx()], ok = fetch()
	println(len(arr[0]), ok)
}

// Output:
// idx
// fetch
// 1 true
