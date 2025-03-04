package main

const (
	A = uint8(^0)
)

func main() {
	println(A)
}

// Error:
// main/files/const55.gno:4:6: bigint underflows target kind
