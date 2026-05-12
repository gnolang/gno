// MAXALLOC: 500000000
package main

// Compares two 10-byte strings with all comparison operators.
//
// large_string_cmp.gno: same test with 10000-byte strings
func main() {
	buf := make([]byte, 10)
	s1 := string(buf)
	s2 := string(buf)
	_ = s1 == s2
	_ = s1 < s2
	_ = s1 <= s2
	_ = s1 > s2
	_ = s1 >= s2
}

// Gas:
// 13095
