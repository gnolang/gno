package main

type nat [1]int

func (n nat) zero(num [1]int) {
	println(num)
}

func (n nat) one(num nat) {
	println(num)
}

func main() {
	var a nat
	a = [1]int{}

	a.zero(nat{0})
	a.one([1]int{1})
}

// Output:
// array[(0 int)]
// (array[(1 int)] main.nat)
