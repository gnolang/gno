package main

import (
	"fmt"
)

const (
	ia = 1_000_000
	ib = 1_00_00_00
	ic = 1_000

	fa = 1_000_000.000
	fb = 1_000_000.000_000

	// Imaginary number aren't supported at the moment
	// ima = 100_000i
	// imb = 1 + 100_000i

	ha = 0x_16_32
	hb = 0b_0110_1001
	ho = 0o_644_755
)

func main() {
	fmt.Printf("%d\n", ia)
	fmt.Printf("%d\n", ib)
	fmt.Printf("%d\n", ic)

	fmt.Printf("%f\n", fa)
	fmt.Printf("%f\n", fb)

	fmt.Printf("%x", ha)
	fmt.Printf("%x", hb)
	fmt.Printf("%x\n", ho)
}

// Output:
// 1000000
// 1000000
// 1000
// 1000000.000000
// 1000000.000000
// 163269349ed
