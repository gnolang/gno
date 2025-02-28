package main

func main() {
	const a float64 = 1.5
	println(int64(a))
}

// Error:
// main/files/convert7a.gno:5:10: cannot convert constant of type Float64Kind to Int64Kind
