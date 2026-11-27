package main

func main() {
	const d uint8 = 0
	println(d - 1)
}

// Error:
// main/const_ovf_sub_uint8.gno:5:10-15: d - 1 (constant -1 of type uint8) overflows uint8

// TypeCheckError:
// main/const_ovf_sub_uint8.gno:5:10: d - 1 (constant -1 of type uint8) overflows uint8
