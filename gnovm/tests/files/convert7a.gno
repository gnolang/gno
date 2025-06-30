package main

func main() {
	const a float64 = 1.5
	println(int64(a))
}

// Error:
// main/convert7a.gno:5:10-18: cannot convert constant of type Float64Kind to Int64Kind

// TypeCheckError:
// main/convert7a.gno:5:16: cannot convert a (constant 1.5 of type float64) to type int64
