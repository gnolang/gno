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
// main/files/const55a.gno:5:2: assignment mismatch: 3 variable(s) but 2 value(s)

// TypeCheckError:
// main/files/const55a.gno:5:8: missing init expr for l
