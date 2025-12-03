package main

func main() {
	expected := `hello`[:]
	a := 1
	println(a == expected) // both typed
}

// Error:
// main/cmp_slice_4.gno:6:10-23: invalid operation: a<VPBlock(1,1)> == expected<VPBlock(1,0)> (mismatched types int and string)

// TypeCheckError:
// main/cmp_slice_4.gno:6:15: invalid operation: a == expected (mismatched types int and string)
