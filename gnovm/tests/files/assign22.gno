package main

func main() {
	m := map[string]int{"a": 1}
	var s int
	var ok bool
	s, ok <<= m["a"]
	println(s, ok)
}

// Error:
// main/assign22.gno:7:2-18: assignment operator <<= requires only one expression on lhs and rhs

// TypeCheckError:
// main/assign22.gno:7:8: assignment operation <<= requires single-valued expressions; main/assign22.gno:4:2: declared and not used: m
