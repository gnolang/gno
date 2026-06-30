// recover25a — negative index into an array must be recoverable, mirroring
// recover25 (positive OOB). Covers the ii<0 branch of the bounds check.

package main

var a [10]int

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	i := -1
	_ = a[i]
}

// Output:
// recover: runtime error: index out of range [-1]
