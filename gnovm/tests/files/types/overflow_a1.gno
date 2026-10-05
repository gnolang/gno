package main

func main() {
	const Huge = 1e1000
	println(Huge)
}

// Error:
// main/overflow_a1.gno:5:2-15: cannot convert untyped bigdec to float64 -- too close to +-Inf

// TypeCheckError:
// main/overflow_a1.gno:5:10: cannot use Huge (untyped float constant 1e+1000) as float64 value in argument to built-in println (overflows)
