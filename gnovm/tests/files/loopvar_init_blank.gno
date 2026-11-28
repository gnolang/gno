package main

// "_" in a three-clause loop's init takes no block slot, so the per-iteration
// copy covers only the names it declares: an empty body ran past the block's
// one slot and panicked, and a variable the body declares was copied as if it
// were an init variable.

func two() (int, int) { return 0, 0 }

func main() {
	for _, x := two(); x < 2; x++ {
	}
	var ps []*[2]int
	for _, x := two(); x < 3; x++ {
		a := [2]int{x}
		ps = append(ps, &a)
	}
	for _, p := range ps {
		print(p[0])
	}
	println()
}

// Output:
// 012
