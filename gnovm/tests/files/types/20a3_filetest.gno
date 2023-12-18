package main

func main() {
	println(int(1) < "a")
}

// Error:
// main/files/types/20a3_filetest.gno:4: cannot use untyped string as IntKind
