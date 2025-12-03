package main

func main() {
	intPtr := new(int8)
	*intPtr = 5

	i := 0
	println(intPtr == &i)
}

// Error:
// main/eql_0a03.gno:8:10-22: invalid operation: intPtr<VPBlock(1,0)> == &(i<VPBlock(1,1)>) (mismatched types *int8 and *int)

// TypeCheckError:
// main/eql_0a03.gno:8:20: invalid operation: intPtr == &i (mismatched types *int8 and *int)
