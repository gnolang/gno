package main

func main() {
	a, b, c := 1, a+1
	println(a)
	println(b)
	println(c)
}

// Error:
// main/var19.gno:4:2-19: assignment mismatch: 3 variable(s) but 2 value(s)

// TypeCheckError:
// main/var19.gno:4:16: undefined: a
