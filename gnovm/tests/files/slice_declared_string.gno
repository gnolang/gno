package main

type Name string

func main() {
	n := Name("_admin")

	// An untyped constant compared with n[0:1] takes its static type,
	// which is Name, as in Go.
	println(n[0:1] == "_")
	println(n[0:1] != "_")
	switch n[0:1] {
	case "_":
		println("reserved")
	default:
		println("not reserved")
	}

	// An explicit conversion turns the Name slice into a string.
	println(string(n[1:]) == "admin")
}

// Output:
// true
// false
// reserved
// true
