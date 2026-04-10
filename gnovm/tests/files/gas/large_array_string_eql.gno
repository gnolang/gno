// MAXALLOC: 500000000
package main

// [10]string with 500-byte strings ==
// small_array_string_eql.gno: [2]string with 1-byte strings
func main() {
	buf := make([]byte, 500)
	s := string(buf)
	a1 := [10]string{s, s, s, s, s, s, s, s, s, s}
	a2 := [10]string{s, s, s, s, s, s, s, s, s, s}
	_ = a1 == a2
}

// Gas:
// 18411
