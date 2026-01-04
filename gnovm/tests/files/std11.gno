package main

import "chain/runtime"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	runtime.CurrentRealm()
}

// Output:
// undefined
