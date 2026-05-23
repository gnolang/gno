// run

// Array of zero-sized elements is itself zero-sized.
// [10]struct{} has size 0 because struct{} has size 0.

package main

func main() {
	p := new([10]struct{})
	q := new([10]struct{})
	if p != q {
		println("FAIL")
		return
	}
	println("ok")
}

// Output:
// ok
