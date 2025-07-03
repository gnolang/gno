package main

import "std"

func main() {
	caller := std.OriginCaller()
	println(caller)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
