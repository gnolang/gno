package main

const (
	d, e = 1, "hello"
	m
)

func main() {
	println(d, e)
	println(m)
}

// Error:
// main/files/const25.gno:5:2: assignment mismatch: 1 variable(s) but 2 value(s)
