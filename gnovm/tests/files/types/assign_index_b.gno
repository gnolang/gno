package main

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	var s int
	var ok int
	s, ok = m["a"]
	println(s)
	println(ok)
}

// Error:
// main/assign_index_b.gno:8:2-16: want bool type got int

// TypeCheckError:
// main/assign_index_b.gno:8:10: cannot use m["a"] (untyped bool value) as int value in assignment
