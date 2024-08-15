package main

func main() {
	foo()[:]
}

func foo() [2]byte {
	return [2]byte{1, 2}
}

// Error:
// doOpSlice: expr not addressable: (array[0x0102] [2]uint8)