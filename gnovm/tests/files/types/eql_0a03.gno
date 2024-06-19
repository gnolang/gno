package main

func main() {
	intPtr := new(int8)
	*intPtr = 5

	i := 0
	println(intPtr == &i)
}

// Error:
// main/files/types/eql_0a03.gno:8: cannot use int8 as int
