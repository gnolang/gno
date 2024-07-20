package main

type nat []int

// package block
var u1, u2 []int = x()

func x() (nat, []int) {
	return []int{1}, nat{2}
}

func main() {
	println(u1)
	println(u2)
}

// Output:
// slice[(1 int)]
// slice[(2 int)]
