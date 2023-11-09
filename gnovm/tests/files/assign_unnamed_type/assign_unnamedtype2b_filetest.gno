package main

type nat [1]int

func (n nat) zero() nat {
	return [1]int{0}
}

func (n nat) one() [1]int {
	return nat{1}
}

func main() {
	var a nat
	a = [1]int{}

	println(a.zero())
	println(a.zero().zero())
	println(a.zero().one())
	println(a.one())
}

// Output:
// (array[(0 int)] main.nat)
// (array[(0 int)] main.nat)
// array[(1 int)]
// array[(1 int)]
