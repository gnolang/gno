package main

type T int

func main() {
	var y struct{ T T }
	_ = (struct{ T })(y)
}

// Error:
// main/convert_struct_embedded.gno:7:6-22: cannot convert y<VPBlock(1,0)> (of type struct{T main.T}) to type struct{T main.T}

// TypeCheckError:
// main/convert_struct_embedded.gno:7:20: cannot convert y (variable of type struct{T T}) to type struct{T}
