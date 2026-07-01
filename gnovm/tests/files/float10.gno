// Issues 12576 and 12621: a negative untyped float constant with small
// magnitude must round to +0, not -0, when typed as float64.

package main

import "math"

var m = -1e-10000

func main() {
	if math.Signbit(m) {
		panic("negative zero")
	}
	println("ok")
}

// Output:
// ok
