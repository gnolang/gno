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
	fn()
}

func fn() {
	u1, n2 = x()
	println(u1)
	println(n2)
	println(u1)
	println(n2)

	// function literal block
	func() {
		u1, n2 = y()
		println(u1)
		println(n2)
		println(u1)
		println(n2)

	}()
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
