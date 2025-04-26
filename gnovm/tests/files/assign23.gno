package main

func main() {
	m := map[string]int{"a": 1}
	var s int
	var ok bool
	s, ok += m["a"]
	println(s, ok)
}

// Error:
// main/files/assign23.gno:7:2: assignment operator += requires only one expression on lhs and rhs

// TypeCheckError:
// main/files/assign23.gno:7:8: assignment operation += requires single-valued expressions; main/files/assign23.gno:4:2: declared and not used: m
