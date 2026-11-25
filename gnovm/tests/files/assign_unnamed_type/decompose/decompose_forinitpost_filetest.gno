package main

// Init and post both convert; the loop sees each fresh result.

type NamedMap map[string]int

var calls int

func fetch() (NamedMap, bool) { calls++; return NamedMap{"a": calls}, calls < 3 }

func main() {
	var m map[string]int
	var ok bool = true
	n := 0
	for m, ok = fetch(); ok; m, ok = fetch() {
		n++
		if n > 10 {
			break
		}
	}
	println(n, m["a"], ok, calls)
}

// Output:
// 2 3 false 3
