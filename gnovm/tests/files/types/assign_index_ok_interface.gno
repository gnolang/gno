package main

func main() {
	m := map[int]int{1: 2}
	var ok error
	_, ok = m[0]
	println(ok)
}

// Error:
// main/assign_index_ok_interface.gno:6:2-14: <untyped> bool does not implement .uverse.error (missing method Error)

// TypeCheckError:
// main/assign_index_ok_interface.gno:6:10: cannot use m[0] (value of type bool) as error value in assignment: bool does not implement error (missing method Error)
