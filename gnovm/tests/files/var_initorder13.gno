package main

// Tests that function call arguments are analysed as init-order constraints.
// A = incr(B) passes B directly as an argument, but the dependency on B is
// created even if B is not used in the body of incr.
// Init order: B (no deps), A (depends on B via argument expression).

var counter int

func incr(arg int) int {
	counter += 1
	return counter
}

var (
	A = incr(B)
	B = incr(0)
)

func main() { println(A) }

// Output:
// 2
