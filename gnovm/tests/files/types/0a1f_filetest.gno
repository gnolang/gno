package main

func main() {
	expected := `hello`[:]
	a := 1
	println(a == expected) // both typed
}

// Error:
// main/files/types/0a1f_filetest.gno:6: invalid operation: mismatched types int and string
