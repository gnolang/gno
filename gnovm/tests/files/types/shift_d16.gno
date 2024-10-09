package main

func main() {
	x := 64
	y := uint64(1<<x - 1)
	println(y)
}

// Output:
// 18446744073709551615
