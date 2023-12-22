package main

func main() {
	r := uint64(1<<2 == 1>>2)
	println(r)
}

// Error:
// main/files/types2/10a001_filetest.gno:4: cannot convert BoolKind to Uint64Kind