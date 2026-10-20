package main

func main() {
	const c1 = 1 << 128
	println(c1)
}

// Error:
// main/overflow16.gno:5:2-13: bigint overflows target kind

// TypeCheckError:
// main/overflow16.gno:5:10: cannot use c1 (untyped int constant 340282366920938463463374607431768211456) as int value in argument to built-in println (overflows)
