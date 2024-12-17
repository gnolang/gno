package main

type nat []int

func main() {
	var a nat
	a = []int{0}
	b := []int{1}

	println(a)
	println(b)

	a = nat{0}
	b = a

	println(a)
	println(b)
}

// Output:
// (slice[(0 int)] main.nat)
// slice[(1 int)]
// (slice[(0 int)] main.nat)
// slice[(0 int)]
