package main

type nat []int

func main() {
	var a nat
	b := any(nat{1})
	a = b.(nat)

	println(b)
}

// Output:
// (slice[(1 int)] main.nat)

// TypeCheckError:
// main/assgin_interface2.gno:6:6: declared and not used: a
