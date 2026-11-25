package main

type Name string

const Prefix Name = "_"

func main() {
	n := Name("_admin")

	// n[0:1] has type Name, as in Go, so it compares with a typed
	// constant of that type and assigns to Name without a conversion.
	println(n[0:1] == Prefix)
	var head Name = n[:1]
	println(head == Prefix)
}

// Output:
// true
// true
