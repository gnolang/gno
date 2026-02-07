package main

func main() {
	expected := `hello`[:]
	a := 1
	println(a == expected) // both typed
}

// Error:
// main/eql_0a1f.gno:6:10-23: invalid operation: (mismatched types int and string)

// TypeCheckError:
// main/eql_0a1f.gno:6:15: invalid operation: a == expected (mismatched types int and string)
