package main

func main() {
	var x int64 = -255

	// Runtime conversion: int8(-255) truncates to int8(1).
	var y int8 = int8(x) // y == 1

	// Confirm y is indeed 1.
	println(y)

	// A clean int8(1) works fine (no prior int64 conversion).
	var clean int8 = 1
	println(1 << clean) // prints 2

	// Shifting by y must also work — y == clean == 1.
	println(1 << y) // prints 2
}

// Output:
// 1
// 2
// 2
