package main

func main() {
	intPtr := new(int)
	*intPtr = 5
	s := "hello"
	println(intPtr == s)
}

// Error:
// main/files/types/0a02.gno:7: *int and string cannot be compared
