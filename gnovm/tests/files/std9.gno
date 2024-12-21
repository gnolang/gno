package main

import "std"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	std.AssertOriginCall()
}

// Output:
// undefined
