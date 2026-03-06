package main

// Range loop: closure captures the range value var, then it's mutated.
// Each iteration has its own per-iteration v; mutation stays local.
func main() {
	var fns []func() int
	for _, v := range []int{1, 2, 3} {
		f := func() int { return v } // capture first
		v += 100                      // mutate after; same iteration's v
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 101
// 102
// 103
