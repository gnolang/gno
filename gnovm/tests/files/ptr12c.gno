// run

// Nested zero-sized array: [5][0]int has size 0 because [0]int has
// size 0. Same spec note as ptr12b: Go's escape analysis may diverge;
// Gno consistently returns zerobase.

package main

func main() {
	p := new([5][0]int)
	q := new([5][0]int)
	if p != q {
		println("FAIL")
		return
	}
	println("ok")
}

// Output:
// ok
