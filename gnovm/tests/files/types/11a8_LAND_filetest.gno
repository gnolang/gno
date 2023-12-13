package main

func main() {
	println(bool(true) && 0) // left typed, right untyped
}

// Error:
// main/files/types/11a8_LAND_filetest.gno:4: cannot use untyped Bigint/Bigdec as BoolKind
