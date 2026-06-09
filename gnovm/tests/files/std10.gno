package main

import (
	"chain/runtime/unsafe"
)

func main() {
	defer func() {
		// assert no panic
		println(recover())
	}()
	println(unsafe.PreviousRealm())
}

// Output:
// UserRealm{ g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm }
// undefined
