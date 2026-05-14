package main

func main() {
	var s int = 2
	var y float32
	{
		y = (1 << s) << s
	}
	println(y)
}

// Error:
// main/shift_deep3d.gno:7:3-20: operator << not defined on: Float32Kind

// TypeCheckError:
// main/shift_deep3d.gno:7:8: invalid operation: shifted operand 1 (type float32) must be integer
