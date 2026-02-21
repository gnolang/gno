package main

func main() {
	s := "a\u263Ab"
	r := []rune(s)
	println("len", len(r))
	println("cap", cap(r))

	b := []byte(s)
	println("len", len(b))
	println("cap", cap(b))
}

// Output:
// len 3
// cap 3
// len 5
// cap 5
