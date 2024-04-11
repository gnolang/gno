package main

func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	var s string
	var ok bool
	s, ok = m["a"]
	println(s)
	println(ok)
}

// Error:
// main/files/types/assign_index.gno:8: cannot use int as string
