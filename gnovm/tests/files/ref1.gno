package main

// ref of a declared type variable.
// xv.TV.T is main.myint, which is correct.

type myint int

func main() {
	var a myint = 42
	b := &a
	println(*b)
}

// Output:
// (42 main.myint)
