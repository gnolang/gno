package main

import "std"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	std.GetCallerAt(0)
}

// Output:
// GetCallerAt requires positive arg
