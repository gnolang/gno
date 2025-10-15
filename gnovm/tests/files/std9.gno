package main

import "chain/runtime"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	runtime.AssertOriginCall()
}

// Output:
// invalid non-origin call
