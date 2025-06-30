package main

func main() {
	println(1 << nil)
}

// Error:
// main/shift_d38.gno:4:10-18: cannot convert (const (undefined)) to type uint

// TypeCheckError:
// main/shift_d38.gno:4:15: cannot convert nil to type uint
