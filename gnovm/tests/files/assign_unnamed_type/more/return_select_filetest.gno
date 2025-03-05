package main

type nat []int

func x() any {
	a := nat{0}
	return a
}

func (n nat) double() nat {
	m := append(n, n...)
	return m
}

func main() {
	a := x()

	b := a.(nat).double().double()
	println(a)
	println(b)
}

// Output:
// (slice[(0 int)] main.nat)
// (slice[(0 int),(0 int),(0 int),(0 int)] main.nat)
