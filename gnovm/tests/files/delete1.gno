package main

type S struct {
	M map[string]int
}

var pkgM map[string]int

func ret() map[string]int { return nil }

// Per the Go spec, delete on a nil map is a no-op.
func main() {
	var m map[string]int
	delete(m, "k")
	delete(pkgM, "k")
	var s S
	delete(s.M, "k")
	delete(ret(), "k")
	delete(map[string]int(nil), "k")
	// An unhashable interface key on a nil map also no-ops: the Go spec
	// says delete on a nil map is a no-op (the gc runtime panics here;
	// gno follows the spec text and its own nil-map read behavior).
	var mi map[interface{}]int
	delete(mi, []int{1})
	println("ok", len(m), len(pkgM), len(s.M), len(mi))
}

// Output:
// ok 0 0 0 0
