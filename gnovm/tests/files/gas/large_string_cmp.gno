// MAXALLOC: 500000000
package main

// Compares two 10000-byte strings with all comparison operators.
//
// small_string_cmp.gno: same test with 10-byte strings
// string_eql_diff_len.gno: equality with different-length strings (early-exit)
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
// 68626
