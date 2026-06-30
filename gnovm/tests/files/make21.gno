package main

// Same compile-time rejection as make20, but for a byte array, which is
// sized via the AllocateDataArray (1 byte/elem) path rather than the
// AllocateListArray (TypedValue/elem) path.

func main() {
	var arr [9223372036854775807]byte
	_ = arr
}

// Error:
// main/make21.gno:8:10-35: type [9223372036854775807]uint8 larger than address space
