package main

func main() {
	x := 11
	var y uint64
	y = -(1 + 1<<x)
	println(y)
}

// Output:
// 18446744073709549567
