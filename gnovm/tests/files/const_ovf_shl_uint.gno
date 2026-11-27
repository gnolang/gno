package main

func main() {
	const k uint = 3
	println(k << 63)
}

// Error:
// main/const_ovf_shl_uint.gno:5:10-17: k << 63 (constant 27670116110564327424 of type uint) overflows uint

// TypeCheckError:
// main/const_ovf_shl_uint.gno:5:10: k << 63 (constant 27670116110564327424 of type uint) overflows uint
