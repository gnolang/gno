package main

const (
	i int32 = 1
	s       = "world"
	c       = 1 << iota

	s2   = "hey"
	d, e = 1, "hello"

	m, n

	b = 1 + 1
	u = +1
)

func main() {
	println(i)
	println(s)
	println(c)
	println(s2)
	println(d, e)
	println(m, n)
	println(b)
	println(u)
}

// Output:
// 1
// world
// 4
// hey
// 1 hello
// 1 hello
// 2
// 1
