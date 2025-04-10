package main

import "chain/runtime"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	runtime.CallerAt(0)
}

// Output:
// CallerAt requires positive arg
