package main

// ref of a concrete int variable.
// xv.TV.T is int, which is correct.

func main() {
	var a int = 42
	b := &a
	println(*b)
}

// Output:
// 42
