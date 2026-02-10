package main

import "chain/runtime"

func main() {
	caller := runtime.OriginCaller()
	println(caller)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
