package main

func main() {
	m := map[string]int{"a": 1}
	var s int
	var ok bool
	s, ok += m["a"]
	println(s, ok)
}

// Error:
// main/assign23.gno:7:2-17: assignment operator += requires only one expression on lhs and rhs

// TypeCheckError:
// main/assign23.gno:7:8: assignment operation += requires single-valued expressions; main/assign23.gno:4:2: declared and not used: m
