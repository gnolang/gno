package main

type nat [1]int

func main() {
	var a nat
	a = [1]int{0}
	b := [1]int{1}

	println(a)
	println(b)
	a = nat{0}
	b = a

	println(a)
	println(b)
}

// Output:
// (array[(0 int)] main.nat)
// array[(1 int)]
// (array[(0 int)] main.nat)
// array[(0 int)]
