// MAXALLOC: 500000000
package main

// Checks equality of two strings with different lengths (10000 vs 5000).
// The length mismatch triggers an early exit with no per-byte gas cost.
//
// large_string_cmp.gno: equal-length 10000-byte strings
func main() {
	buf1 := make([]byte, 10000)
	buf2 := make([]byte, 5000)
	s1 := string(buf1)
	s2 := string(buf2)
	_ = s1 == s2
}

// Gas:
// 19065
