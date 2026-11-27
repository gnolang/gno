package main

func main() {
	const a uint32 = 3000000000
	println(a + a)
}

// Error:
// main/const_ovf_add_uint32.gno:5:10-15: a + a (constant 6000000000 of type uint32) overflows uint32

// TypeCheckError:
// main/const_ovf_add_uint32.gno:5:10: a + a (constant 6000000000 of type uint32) overflows uint32
