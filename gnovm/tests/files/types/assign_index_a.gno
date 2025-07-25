package main

func main() {
	var s string
	var ok bool
	s, ok = map[string]int{"a": 1, "b": 2, "c": 3}["a"]
	println(s)
	println(ok)
}

// Error:
// main/assign_index_a.gno:6:2-53: cannot use int as string

// TypeCheckError:
// main/assign_index_a.gno:6:10: cannot use map[string]int{…}["a"] (map index expression of type int) as string value in assignment
