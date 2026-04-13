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
// main/const55.gno:10:10-11: name not defined: m

// TypeCheckError:
// main/const55.gno:5:2: extra init expr at main/const55.gno:4:2
