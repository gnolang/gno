package main

// Tests that a method call through a package-level receiver variable with a
// pointer receiver is correctly ordered. t.GetB() calls (*T).GetB which reads
// B, so A's dependency on B is inferred via the method body.
// Also checks that t (the receiver variable) is itself treated as a dep of A,
// and t has no deps, so init order is: T (type), t (no deps), B (no deps), A.

type T struct{}

func (*T) GetB() int { return B }

var t T

var (
	A = t.GetB()
	B = 42
)

func main() { println(A) }

// Output:
// 42
