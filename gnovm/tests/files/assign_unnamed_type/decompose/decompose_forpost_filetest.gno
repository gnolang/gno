package main

// The converting assignment in a for-post runs after every iteration,
// including one left by continue.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

func main() {
	var m map[string]int
	var ok bool
	n := 0
	for ; n < 2; m, ok = fetch() {
		n++
		println("iter", n, len(m), ok)
		if n == 1 {
			continue
		}
	}
	println("done", len(m), ok)
}

// Output:
// iter 1 0 false
// iter 2 1 true
// done 1 true
