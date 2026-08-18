package main

// Multiple for-init vars: both i and j are loop vars (NumInit==2).
// Shadow them to get per-iteration snapshots without body mutation.
func main() {
	var fns []func() int
	for i, j := 0, 10; i < 3; i++ {
		i := i
		j := j
		fns = append(fns, func() int { return i*100 + j })
		j += 5 // modifies body-scope j after snapshot; closure sees it (j=15)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 15
// 115
// 215
