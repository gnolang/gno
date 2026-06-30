// v, ok = m[k] over a declared (named) map type: value type must be checked
// against the base map's value type.
package main

type M map[string]int

func main() {
	m := M{"a": 1}
	var v bool
	var ok bool
	v, ok = m["a"]
	println(v, ok)
}

// Error:
// main/assign_index_d.gno:11:2-16: cannot use int as bool

// TypeCheckError:
// main/assign_index_d.gno:11:10: cannot use m["a"] (map index expression of type int) as bool value in assignment
