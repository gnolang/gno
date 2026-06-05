package main

type T int

func main() {
	var x struct{ T }
	var y struct{ T T }
	x = y
	_ = x
}

// Error:
// main/assign_struct_embedded.gno:8:2-7: cannot use struct{T main.T} as struct{T main.T}

// TypeCheckError:
// main/assign_struct_embedded.gno:8:6: cannot use y (variable of type struct{T T}) as struct{T} value in assignment
