package main

const (
	d, e = 1, "hello"
	m, n, l
)

func main() {
	println(d, e)
	println(m)
}

// Error:
// main/const55a.gno:10:10-11: name not defined: m

// TypeCheckError:
// main/const55a.gno:5:8: missing init expr for l
