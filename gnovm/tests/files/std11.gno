package main

import "chain/runtime"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	runtime.CallerAt(42)
}

// Output:
// frame not found
