package main

func main() {
	const a int8 = -64
	println(-(a + a))
}

// Error:
// main/const_ovf_neg_sum_int8.gno:5:10-18: -(a + a) (constant 128 of type int8) overflows int8

// TypeCheckError:
// main/const_ovf_neg_sum_int8.gno:5:10: -(a + a) (constant 128 of type int8) overflows int8
