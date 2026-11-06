package main

// len of an array variable is a constant, so a type can depend on a
// variable's type. x's declaration needs T built, not just reserved.

var x [3]T

type E [len(x)]int

type T int

func main() {
	var e E
	println(len(e))
}

// Output:
// 3
