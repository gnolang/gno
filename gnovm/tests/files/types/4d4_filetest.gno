package main

// both untyped const
// TODO: dec value representation
func main() {
	println(1.0 / 0)
}

// Error:
// main/files/types/4d4_filetest.gno:6: bigdec division error: division by zero
