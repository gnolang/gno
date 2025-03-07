package main

func main() {
	const f = 0x1p1023 * (1 + (1 - 0x1p-52))
	println(float32(f))
}

// Error:
// main/files/const62.gno:5:10: cannot convert constant of type Float64Kind to Float32Kind
