package main

type nat []int

var n1, n2 nat = x()

func x() (nat, []int) {
	return nat{1}, nat{2}
}

func main() {
	println(n1)
	println(n2)
}

// Output:
// (slice[(1 int)] main.nat)
// (slice[(2 int)] main.nat)
