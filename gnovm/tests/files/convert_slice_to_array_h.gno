package main

func main() {
	s := []byte{1, 2, 3}
	a := [0]byte(s)
	println(len(a))
}

// Output:
// 0
