package main

import (
	"chain/runtime/unsafe"
)

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	unsafe.CurrentRealm()
}

// Output:
// undefined
