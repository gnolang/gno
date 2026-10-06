package main

// Order:
// B has no deps        -> initialized first (incr->counter=1, B=1).
// A has no deps        -> initialized second (incr->counter=2, A=2).
// C depends on B and A -> initialized last (C=B+A=3).

var counter int

func incr() int { counter++; return counter }

var (
	C = B + A
	B = incr()
	A = incr()
)

func main() { println(A, B, C) }

// Output:
// 2 1 3
