package main

func main() {
	var x any = 2
	var ok error
	_, ok = x.(int)
	println(ok)
}

// Error:
// main/assign_type_assertion_ok_interface.gno:6:2-17: <untyped> bool does not implement .uverse.error (missing method Error)

// TypeCheckError:
// main/assign_type_assertion_ok_interface.gno:6:10: cannot use x.(int) (value of type bool) as error value in assignment: bool does not implement error (missing method Error)
