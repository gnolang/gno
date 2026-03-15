package main

// Tests that function call arguments are analysed as init-order constraints.
// A = double(B) passes B directly as an argument; without queuing CallExpr.Args,
// findDependentNames would not discover the dependency on B and could initialize
// A before B, producing double(0)=0 instead of double(21)=42.
// Init order: B (no deps), A (depends on B via argument expression).

func double(x int) int { return x * 2 }

var (
	A = double(B)
	B = 21
)

func main() { println(A) }

// Output:
// 42
