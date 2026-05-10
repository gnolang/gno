package main

// Drains gas via byte-array equality on the Data fast path.
// Verifies the per-byte metering kicks in for [N]byte (previously
// unmetered when the Data fast path returned bytes.Equal directly).
//
// large_array_eql.gno: same pattern with [100]int (per-element loop)
// large_array_string_eql.gno: same pattern with [10]string
func main() {
	a1 := [100]byte{}
	a2 := [100]byte{}
	_ = a1 == a2
}

// Gas:
// 3941
