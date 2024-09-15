package main

type nat []int

func main() {
	var a nat
	b := interface{}(nat{1})
	a = b.(nat)

	println(b)
}

// Output:
// (slice[(1 int)] main.nat)
