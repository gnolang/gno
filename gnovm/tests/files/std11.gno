package main

import "std"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	std.CallerAt(42)
}

// Output:
// frame not found
