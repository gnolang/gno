// run

// Array of zero-sized elements is itself zero-sized per the spec:
// [10]struct{} has size 0 because struct{} has size 0. In Go, escape
// analysis may or may not route these through runtime.zerobase, so
// == is non-deterministic. Gno consistently returns zerobase.

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
