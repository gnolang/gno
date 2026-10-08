package main

// Same through a composite literal: len(T{}) is constant only once T is
// built.

type E [len(T{})]int

type T [3]int

func main() {
	var e E
	println(len(e))
}

// Output:
// 3
