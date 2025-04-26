package main

func main() {
	expected := `hello`[:]
	a := 1
	println(a == expected) // both typed
}

// Error:
// main/files/types/eql_0a1f.gno:6:10: cannot use int as string

// TypeCheckError:
// main/files/types/eql_0a1f.gno:6:15: invalid operation: a == expected (mismatched types int and string)
