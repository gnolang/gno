package main

func main() {
	var x *struct{ A int `a` }
	var y *struct{ A int `b` }
	x = y
	_ = x
}

// Error:
// main/assign_ptr_struct_tags.gno:6:2-7: cannot use *struct{A int} as *struct{A int}

// TypeCheckError:
// main/assign_ptr_struct_tags.gno:6:6: cannot use y (variable of type *struct{A int "b"}) as *struct{A int "a"} value in assignment
