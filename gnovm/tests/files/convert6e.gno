package main

func main() {
	const a float32 = 1.5
	println(int32(a))
}

// Error:
// main/files/convert6e.gno:5:10: cannot convert constant of type Float32Kind to Int32Kind
