package main

type nat []int

func main() {
	var a nat
	b := any(nat{1})
	a = b.(nat)

	println(b)
}

// TypeCheckError:
// main/assgin_interface2.gno:6:6: declared and not used: a

// Error:
// main/assgin_interface2.gno:6:6-11': declared and not used: a
