package main

type nat []int

var n1, u2 = x()

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

	// multiple statements
	println(n1)
	println(u2)

	u2, n1 = y()
	println(n1)
	println(u2)

	n1, u2 = x()
	println(n1)
	println(u2)

}

// Output:
// (slice[(1 int)] main.nat)
// slice[(2 int)]
// (slice[(4 int)] main.nat)
// slice[(3 int)]
// (slice[(1 int)] main.nat)
// slice[(2 int)]
