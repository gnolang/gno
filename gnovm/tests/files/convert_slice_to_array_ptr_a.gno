// Basic (*[N]T)(s): pointer aliases the slice's underlying array.
package main

func main() {
	s := []byte{1, 2, 3, 4, 5}
	p := (*[3]byte)(s)
	(*p)[0] = 99
	println(s[0], (*p)[0], s[1], (*p)[1], s[2], (*p)[2])
}

// Output:
// 99 99 2 2 3 3
