package main

const (
	A = uint(^1)
)

func main() {
	println(A)
}

// Error:
// main/const60.gno:4:6-14: bigint underflows target kind

// TypeCheckError:
// main/const60.gno:4:11: constant -2 overflows uint
