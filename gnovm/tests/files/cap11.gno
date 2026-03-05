package main

func main() {
	s := "a\u263Ab"
	r := []rune(s)
	println("len", len(r))
	println("cap", cap(r))

	b := []byte(s)
	println("len", len(b))
	println("cap", cap(b))

	// empty string
	re := []rune("")
	println("len", len(re))
	println("cap", cap(re))

	be := []byte("")
	println("len", len(be))
	println("cap", cap(be))

	// all non-ASCII runes
	s2 := "你好"
	r2 := []rune(s2)
	println("len", len(r2))
	println("cap", cap(r2))

	b2 := []byte(s2)
	println("len", len(b2))
	println("cap", cap(b2))
}

// Output:
// len 3
// cap 3
// len 5
// cap 5
// len 0
// cap 0
// len 0
// cap 0
// len 2
// cap 2
// len 6
// cap 6
