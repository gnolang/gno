package main

const (
	A = uint(^1)
)

func main() {
	println(A)
}

// Error:
// main/files/const60.gno:4:6: bigint underflows target kind
