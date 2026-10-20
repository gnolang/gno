package main

func main() {
	const Huge = 1e1000
	println(Huge)
}

// Error:
// main/overflow_a1.gno:5:2-15: cannot convert untyped bigdec to float64 -- too close to +-Inf
