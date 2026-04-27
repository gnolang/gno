// spec: s1 := (*[1]byte)(s[1:]) // &s1[0] == &s[1]
package main

func main() {
	s := []byte{10, 20, 30, 40}
	p := (*[2]byte)(s[1:])
	(*p)[0] = 77
	println(s[1], s[2], (*p)[0], (*p)[1])
}

// Output:
// 77 30 77 30
