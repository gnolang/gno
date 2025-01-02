package main

import "std"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	std.CallerAt(0)
}

// Output:
// CallerAt requires positive arg
