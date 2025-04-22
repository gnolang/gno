package main

func main() {
	println(1 << nil)
}

// Error:
// main/files/types/shift_d38.gno:4:10: cannot convert (const (undefined)) to type uint

// TypeCheckError:
// main/files/types/shift_d38.gno:4:15: cannot convert nil to type uint
