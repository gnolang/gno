package main

func main() {
	s := []byte{1, 2, 3}
	a := [5]byte(s)
	println(a)
}

// Error:
// runtime error: cannot convert slice with length 3 to array with length 5
