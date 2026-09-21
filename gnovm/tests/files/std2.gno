package main

import (
	"chain/runtime/unsafe"
)

func main() {
	caller := unsafe.OriginCaller()
	println(caller)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
