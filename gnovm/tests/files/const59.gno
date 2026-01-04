package main

const (
	A = ^0
	B = ^1
	C = int8(^1)
	D = ^uint8(1)
)

func main() {
	println(A, B, C, D)
}

// Output:
// -1 -2 -2 254
