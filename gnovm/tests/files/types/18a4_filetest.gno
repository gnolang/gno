package main

func main() {
	r := "a"
	r &= "b"
	println(r)
}

// Error:
// main/files/types/18a4_filetest.gno:5: operator &= not defined on: StringKind
