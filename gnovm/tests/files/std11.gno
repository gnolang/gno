package main

import "std"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	std.CurrentRealm()
}

// Output:
// undefined
