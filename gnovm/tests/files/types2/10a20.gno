package main

func main() {
	x := 63
	y := int64(1<<x - 1)
	println(y)
}

// Output:
// 9223372036854775807
