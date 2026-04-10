// MAXALLOC: 500000000
package main

// String == with different lengths (10000 vs 5000); early-exit, no per-byte gas
// large_string_cmp.gno: equal-length 10000-byte strings
func main() {
	buf1 := make([]byte, 10000)
	buf2 := make([]byte, 5000)
	s1 := string(buf1)
	s2 := string(buf2)
	_ = s1 == s2
}

// Gas:
// 39781
