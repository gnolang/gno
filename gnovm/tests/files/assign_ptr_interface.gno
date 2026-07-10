package main

func main() {
	var x *interface{}
	var y *int
	x = y
	_ = x
}

// Error:
// main/assign_ptr_interface.gno:6:2-7: cannot use *int as *interface {}

// TypeCheckError:
// main/assign_ptr_interface.gno:6:6: cannot use y (variable of type *int) as *interface{} value in assignment: *int does not implement *interface{} (type *interface{} is pointer to interface, not interface)
