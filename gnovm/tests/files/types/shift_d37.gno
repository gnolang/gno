package main

func main() {
	var a = int(1) + 1
	println(nil << a)
}

// Error:
// main/shift_d37.gno:5:10-18: operator << not defined on: nil

// TypeCheckError:
// main/shift_d37.gno:5:10: invalid operation: shifted operand nil must be integer
