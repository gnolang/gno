// MAXALLOC: 500000000
package main

// Two slices that do not move a for loop's init variable to the heap, so
// neither loop copies its [1000]int per iteration: the first slices an array
// behind a pointer field of the init variable, which lives in the pointee;
// the second slices an array the body declares, which each iteration declares
// afresh.

type T struct{ buf [2]int }

type S struct {
	big [1000]int
	p   *T
}

func main() {
	n := 0
	for s := (S{p: &T{}}); s.big[0] < 10; s.big[0]++ {
		n += len(s.p.buf[:])
	}
	for i := 0; i < 10; i++ {
		var b [1000]int
		n += len(b[:])
	}
	println(n)
}

// Output:
// 10020

// Gas:
// 1142858
