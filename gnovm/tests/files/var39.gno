// Multi-value decl where each name has a deeper independent chain.
// A -> C -> E, B -> D -> F — two separate chains, no false cycle.
package main

var (
	A, B = C, D
	C    = E
	D    = F
	E    = 10
	F    = 20
)

func main() {
	println(A, B) // 10 20
}

// Output:
// 10 20
