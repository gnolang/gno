package main

// Modifying the loop init var in the body affects the condition check,
// causing early termination. The closure captures the value after the modification.
func main() {
	var fns []func() int
	for i := 0; i < 10; i++ {
		i += 5 // skip ahead; within same iteration
		fns = append(fns, func() int { return i })
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 5
// 11
