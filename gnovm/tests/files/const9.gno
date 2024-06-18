package main

const (
	a = 2
	b = c + d
	c = a + d
	d = e + f
	e = b + 2
	f = 4
)

func main() {
	println(b)
}

// Error:
// main/files/const9.gno:5: constant definition loop with b
