package main

func main() {
	const a int8 = 50
	println((a + a) * a)
}

// Error:
// main/const_ovf_paren_int8.gno:5:10-21: (a + a) * a (constant 5000 of type int8) overflows int8

// TypeCheckError:
// main/const_ovf_paren_int8.gno:5:10: (a + a) * a (constant 5000 of type int8) overflows int8
