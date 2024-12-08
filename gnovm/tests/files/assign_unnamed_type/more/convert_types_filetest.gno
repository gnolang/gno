package main

type nat []int

func main() {
	var b []int
	b = nat{1}
	println(b)

	var a nat
	a = []int{0}
	println(a)

	b = a
	println(b)
}

// Output:
// slice[(1 int)]
// (slice[(0 int)] main.nat)
// slice[(0 int)]
