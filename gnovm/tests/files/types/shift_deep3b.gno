package main

var sink int

func main() {
	var s int = 1
	var y int
	{
		y = 1.2 << s 
	}
	sink = y
	println(sink)
}

// Error:
// main/shift_deep3b.gno:9:7-15: invalid operation: shifted operand (const (1.2 <untyped> bigdec)) (<untyped> bigdec) must be integer

// TypeCheckError:
// main/shift_deep3b.gno:9:7: invalid operation: shifted operand 1.2 (untyped float constant) must be integer
