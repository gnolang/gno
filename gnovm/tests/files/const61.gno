package main

const (
	A = uint8(^0)
)

func main() {
	println(A)
}

// Error:
// main/files/const61.gno:4:6: bigint underflows target kind

// TypeCheckError:
// main/files/const61.gno:4:12: constant -1 overflows uint8
