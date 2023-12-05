package main

type nat []int

func main() {
	var a nat
	var b []int

	a = []int{0}
	b = nat{1}

	println(a)
	println(b)

	a = b
	b = a

	println(a)
	println(b)
}

// Output:
// (slice[(0 int)] main.nat)
// slice[(1 int)]
// (slice[(1 int)] main.nat)
// slice[(1 int)]
