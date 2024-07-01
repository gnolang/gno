package main

func main() {
	println(1 << nil)
}

// Error:
// main/files/types2/shift_d38.gno:3: cannot convert (const (undefined) nil) to UintKind
