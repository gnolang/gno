// MAXALLOC: 500000000
package main

// Checks equality of two structs with 20 int fields.
//
// small_struct_eql.gno: same test with 3 fields
// string_struct_eql.gno: struct with a string field
type Large struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9          int
	F10, F11, F12, F13, F14, F15, F16, F17, F18, F19 int
}

func main() {
	s1 := Large{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	s2 := Large{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	_ = s1 == s2
}

// Gas:
// 13737
