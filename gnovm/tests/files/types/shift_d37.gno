package main

func main() {
	var a = int(1) + 1
	println(nil << a)
}

// Error:
// main/shift_d37.gno:5:10-18: invalid operation: shifted operand (const (undefined)) (<nil>) must be integer

// TypeCheckError:
// main/shift_d37.gno:5:10: invalid operation: shifted operand nil must be integer
