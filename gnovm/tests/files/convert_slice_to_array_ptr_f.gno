// Named element type: identical TypeID on both sides.
package main

type B byte

func main() {
	s := []B{1, 2, 3, 4}
	p := (*[3]B)(s[1:])
	(*p)[0] = 99
	println(s[0], s[1], s[2], s[3])
}

// Output:
// (1 main.B) (99 main.B) (3 main.B) (4 main.B)
