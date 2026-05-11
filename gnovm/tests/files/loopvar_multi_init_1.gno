package main

// Multiple for-init vars directly captured by closures (no shadowing).
// Exercises NumInit==2 heap-copy path. j is mutated in the body;
// the closure for each iteration sees j's within-iteration modified value.
// i is not mutated in the body; closure sees pre-post-stmt value (0, 1, 2).
func main() {
	var fns []func() (int, int)
	for i, j := 0, 10; i < 3; i++ {
		f := func() (int, int) { return i, j }
		fns = append(fns, f)
		j += 5 // modify loop var j in body
	}
	for _, fn := range fns {
		a, b := fn()
		println(a, b)
	}
}

// Output:
// 0 15
// 1 20
// 2 25
