package main

const (
	d, e = "hello", 1
	m, n
	a string = "a"
	k        = 1 << iota
)

func main() {
	println(d, e)
	println(m, n)
	println(k)
}

// Output:
// hello 1
// hello 1
// 8
