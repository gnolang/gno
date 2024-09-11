package main

func main() {
	expected := `hello`[:]
	a := 1
	println(a == expected) // both typed
}

// Error:
// main/files/types/cmp_slice_4.gno:6:10: cannot use int as string
