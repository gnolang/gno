package main

type nat []int

// gloabal variables
var u1 []int
var n2 nat

func x() (nat, []int) {
	a := nat{1}
	b := []int{2}
	return a, b
}
func y() (nat, []int) {
	a := nat{3}
	b := []int{4}
	return a, b
}
func main() {

	// else-if, for block
	if false {

	} else if true {
		u1, n2 = x()
		println(u1)
		println(n2)
		println(u1)
		println(n2)
		u1, n2 = y()
		println(u1)
		println(n2)
		println(u1)
		println(n2)
		u1, n2 = x()
		println(u1)
		println(n2)
		println(u1)
		println(n2)

		for i := 0; i < 2; i++ {
			u1, n2 = y()
			println(u1)
			println(n2)
			println(u1)
			println(n2)
		}
	}
}

// Output:
// slice[(1 int)]
// (slice[(2 int)] main.nat)
// slice[(1 int)]
// (slice[(2 int)] main.nat)
// slice[(3 int)]
// (slice[(4 int)] main.nat)
// slice[(3 int)]
// (slice[(4 int)] main.nat)
// slice[(1 int)]
// (slice[(2 int)] main.nat)
// slice[(1 int)]
// (slice[(2 int)] main.nat)
// slice[(3 int)]
// (slice[(4 int)] main.nat)
// slice[(3 int)]
// (slice[(4 int)] main.nat)
// slice[(3 int)]
// (slice[(4 int)] main.nat)
// slice[(3 int)]
// (slice[(4 int)] main.nat)
