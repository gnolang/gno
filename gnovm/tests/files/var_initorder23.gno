package main

// Tests that multiple blank-identifier declarations are initialized in
// source order and their side effects are visible to later declarations.
// Blank decls produce no names in fdeclared but still occupy their
// position in the pending list.

var counter int

func incr() int { counter++; return counter }

var (
	_ = incr() // counter=1
	_ = incr() // counter=2
	A = counter // should be 2
)

func main() { println(A) }

// Output:
// 2
