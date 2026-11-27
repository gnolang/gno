package main

func main() {
	const a byte = 200
	println(a + a)
}

// Error:
// main/const_ovf_add_byte.gno:5:10-15: a + a (constant 400 of type uint8) overflows uint8

// TypeCheckError:
// main/const_ovf_add_byte.gno:5:10: a + a (constant 400 of type byte) overflows byte
