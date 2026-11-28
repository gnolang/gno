// MAXALLOC: 1150000
package main

type Leaf [100000]byte

type Pair struct {
	A, B Leaf
}

type Nested [2]Pair

type Box struct {
	V any
}

var a Nested

func t() *Box { return &Box{} }

// Regression: each target's copy runs one leaf at a time, and a leaf's
// allocation can run a collection while earlier leaves sit in the buffers
// ArrayValue.Copy and StructValue.Copy are still filling. That collection must
// count them. The cap sits above what the program holds with either buffer left
// out of the count, and below what it holds with both counted.
func main() {
	t().V, t().V = a, a
	println("survived")
}

// Error:
// allocation limit exceeded
