package main

var sink int

func main() {
	var s int = 1
	var y int
	// build deep left-leaning chain: (((y<<s)<<s)<<s)...
	{
		y = (1.0 << s) << s 
	}
	sink = y
	println(sink)
}

// TypeCheckError:
// main/shift_deep3.gno:10:7: invalid operation: shifted operand (1.0 << s) (untyped float value) must be integer

// Error:
// main/shift_deep3.gno:10:8-16: operator << not defined on: Float64Kind
