package main

import "encoding/binary"

func main() {
	// Verify that addressable results of expressions are
	// still addressable when accessed via a selector.
	var b []byte
	le := &binary.LittleEndian
	println(&le.AppendUint16(b, 0)[0])
}

// Output:
// &(0 <databyte> uint8)
