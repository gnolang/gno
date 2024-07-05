package main

import (
	"math"
)

func main() {
	// test float64
	f := float64(0.3)
	x := math.Float64bits(f)
	e := uint(40)
	println(f, x, e, (1 << (64 - e)))

	// test int64
	println(x &^ 0xFF)
	x &^= 1<<(64-e) - 1
	println(f, x, e)
}

// Output:
// 0.3 4599075939470750515 40 16777216
// 4599075939470750464
// 0.3 4599075939467395072 40
