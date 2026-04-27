package main

func main() {
	s := []byte{10, 20, 30, 40}
	a := [2]byte(s[1:])
	println(a[0], a[1])
}

// Output:
// 20 30
