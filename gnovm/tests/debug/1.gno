package main

func main() {
	println(int(1) + int8(1))
}

// Go_Error:
// debug/1.gno:8:10: invalid operation: int(1) + int8(1) (mismatched types int and int8)

// Gno_Error:

// Gno_Output:
// 2
