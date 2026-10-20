// MAXALLOC: 500000000
package main

// Gas of a loop over the 256-bit operators; int_ops_baseline.gno runs the
// same loop on int.
func main() {
	var a, b uint256 = 1<<200 + 12345, 1<<100 + 7
	var s int256 = -1 << 150
	for i := 0; i < 100; i++ {
		a = a*3 + b
		b = a/b + a%b
		s = (s >> 1) ^ int256(b)
		_ = a < b
	}
}

// Gas:
// 624648
