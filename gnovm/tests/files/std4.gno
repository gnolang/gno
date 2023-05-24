package main

import (
	"std"
)

func main() {
	caller1 := std.CallerAt(1)
	println(caller1)
}

// Output:
// g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
