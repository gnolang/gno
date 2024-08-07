package main

func main() {
	println(1 << -1)
}

// Error:
// main/files/types/shift_e0.gno:4:10: invalid operation: negative shift count: (const (-1 <untyped> bigint))
