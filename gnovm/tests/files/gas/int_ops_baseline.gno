// MAXALLOC: 500000000
package main

// The loop of int256_ops.gno, on int.
func main() {
	var a, b uint = 1<<60 + 12345, 1<<30 + 7
	var s int = -1
	s <<= 50
	for i := 0; i < 100; i++ {
		a = a*3 + b
		b = a/b + a%b
		s = (s >> 1) ^ int(b)
		_ = a < b
	}
}

// Gas:
// 472004
