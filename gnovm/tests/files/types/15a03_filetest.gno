package main

func main() {
	a := 10
	a %= "a"
	println(a)
}

// Error:
// main/files/types/15a03_filetest.gno:5: cannot use untyped string as IntKind
