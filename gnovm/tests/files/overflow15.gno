package main

func main() {
	const c1 = 1 << 128
	println(uint64(c1 >> 65))
}

// Output:
// 9223372036854775808
