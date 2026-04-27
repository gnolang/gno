// Data-backed byte slice fast path: []byte(string) produces a Data-backed array.
package main

func main() {
	s := []byte("abcdef")
	p := (*[4]byte)(s[1:])
	println(string((*p)[:]))
}

// Output:
// bcde
