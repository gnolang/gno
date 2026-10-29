package main

import "fmt"

// No context: the kind's default type applies to the whole tree.
func main() {
	var s uint
	x := 1 + ('a' << s)
	fmt.Printf("%v %T\n", x, x)
	y := (1 << s) + 'a'
	fmt.Printf("%v %T\n", y, y)
	z := ('a' << s) * 2
	fmt.Printf("%v %T\n", z, z)
	w := (1 << s) + (1 << 40)
	fmt.Printf("%v %T\n", w, w)
	u := -('a' << s)
	fmt.Printf("%v %T\n", u, u)
	println(1+('a'<<s), ('a'<<s)*2, (1<<s)+(1<<40))
	s = 32
	t := ('a' << s) + ('b' << s)
	fmt.Printf("%v %T\n", t, t)
	p := (1 << s) + (1 << s)
	fmt.Printf("%v %T\n", p, p)
	println(('a'<<s)+('b'<<s), (1<<s)+(1<<s))
}

// Output:
// 98 int32
// 98 int32
// 194 int32
// 1099511627777 int
// -97 int32
// 98 194 1099511627777
// 0 int32
// 8589934592 int
// 0 8589934592
