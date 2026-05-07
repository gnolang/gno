// MAXALLOC: 500000000
package main

// Checks equality of two [2]string arrays with 1-byte strings.
//
// large_array_string_eql.gno: [10]string with 500-byte strings
func main() {
	a1 := [2]string{"a", "b"}
	a2 := [2]string{"a", "b"}
	_ = a1 == a2
}

// Gas:
// 5688
