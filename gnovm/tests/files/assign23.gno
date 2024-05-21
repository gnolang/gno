package main

func main() {
	m := map[string]int{"a": 1}
	var s int
	var ok bool
	s, ok += m["a"]
	println(s, ok)
}

// Error:
// main/files/assign23.gno:7: assignment operator += requires only one expression on lhs and rhs
