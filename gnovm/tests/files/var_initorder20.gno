package main

// Tests that init-order analysis correctly handles promoted methods from
// embedded structs. Outer embeds Inner, and Inner.GetB reads B. The call
// Outer{}.GetB() should be traced through to Inner.GetB's body, discovering
// the transitive dependency on B.

type Inner struct{}

func (Inner) GetB() int { return B }

type Outer struct{ Inner }

var (
	A = Outer{}.GetB()
	B = 42
)

func main() { println(A) }

// Output:
// 42
