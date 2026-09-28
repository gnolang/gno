// MAXALLOC: 1000000
package main

// Regression: an allocator buffer under construction must stay counted.
//
// `_ = expr` assigns through Block.GetBlankRef, so the block pins expr for
// its whole life even though nothing can ever read it back. Block's
// VisitAssociated walked only Values and Parent, so the value stayed live on
// the Go heap while invisible to Reset()+Recount(): one blank assignment per
// recursion level held 60MB under a 1MB cap, while the same expression bound
// to a name was correctly refused.
func consume(n int) {
	if n == 0 {
		return
	}
	_ = make([]byte, 300_000)
	consume(n - 1)
}

func main() {
	consume(200)
	println("survived")
}

// Error:
// allocation limit exceeded
