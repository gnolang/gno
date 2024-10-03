package main

func main() {
	intPtr := new(int)
	*intPtr = 5
	s := "hello"
	println(intPtr == s)
}

// Error:
// main/files/types/eql_0a02.gno:7:10: cannot use *int as string
