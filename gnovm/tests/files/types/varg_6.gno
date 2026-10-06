package main

func bar(s ...any) {
	var s1 []any
	s1 = append(s1, s...)
	println(s1)
}
func main() {
	var a uint = 2
	bar(1.0<<a, 2)
}

// Error:
// main/varg_6.gno:10:2-16: operator << not defined on: Float64Kind

// TypeCheckError:
// main/varg_6.gno:10:6: invalid operation: shifted operand 1.0 (type float64) must be integer
