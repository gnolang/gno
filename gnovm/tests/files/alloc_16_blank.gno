// MAXALLOC: 1000000
package main

// Regression: a value assigned to "_" must not stay held.
//
// `_ = expr` once stored expr through Block.GetBlankRef, so the block pinned
// it for its whole life even though nothing can ever read it back, and a
// collection that skipped the blank slot left it uncounted. Nothing reads
// "_", so the assignment now stores nothing: each recursion level's 300KB is
// garbage a collection frees, as in Go, and the program completes.
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

// Output:
// survived
