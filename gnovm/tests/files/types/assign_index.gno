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
// main/files/types/assign_index.gno:8:2: cannot use int as string

// TypeCheckError:
// main/files/types/assign_index.gno:8:10: cannot use m["a"] (map index expression of type int) as string value in assignment
