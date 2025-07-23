package main

func main() {
	intPtr := new(int)
	*intPtr = 5
	s := "hello"
	println(intPtr == s)
}

// Error:
// main/eql_0a02.gno:7:10-21: cannot use *int as string

// TypeCheckError:
// main/eql_0a02.gno:7:20: invalid operation: intPtr == s (mismatched types *int and string)
