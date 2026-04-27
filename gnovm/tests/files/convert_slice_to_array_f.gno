package main

func main() {
	s := make([]byte, 2, 4)
	a := [3]byte(s)
	println(a)
}

// Error:
// runtime error: cannot convert slice with length 2 to array with length 3
