package main

type nat []int

func (n nat) zero(num []int) {
	println(num)
}

func (n nat) one(num nat) {
	println(num)
}

func main() {
	var a nat
	a = []int{}

	a.zero(nat{0})
	a.one([]int{1})
}

// Output:
// slice[(0 int)]
// (slice[(1 int)] main.nat)
