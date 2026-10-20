package main

const (
	A = uint8(^0)
)

func main() {
	println(A)
}

// Error:
// main/const61.gno:4:6-15: bigint underflows target kind
