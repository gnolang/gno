package main

func main() {
	s := []byte("abcdef")
	a := [4]byte(s[1:])
	println(string(a[:]))
}

// Output:
// bcde
