package main

func main() {
	var a, b, c = 1, a+1
	println(a)
	println(b)
	println(c)
}

// Error:
// main/files/var19.gno:4:6: assignment mismatch: 3 variable(s) but 2 value(s)
