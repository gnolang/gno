// MAXALLOC: 500000000
package main

// Compare two large strings (10000 bytes) using all comparison
// operators. This test verifies that per-byte gas metering is
// applied to string comparisons (==, <, <=, >, >=).

func main() {
	buf := make([]byte, 10000)
	s1 := string(buf)
	s2 := string(buf)
	_ = s1 == s2
	_ = s1 < s2
	_ = s1 <= s2
	_ = s1 > s2
	_ = s1 >= s2
}

// Gas:
// 2419259
