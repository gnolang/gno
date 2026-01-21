package main

// Slice to array conversion with named array type.

type Hash [4]byte

func main() {
	s := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	h := Hash(s)
	println(h[0], h[1], h[2], h[3])
}

// Output:
// 222 173 190 239
