package main

var a int = 0
var b int = 0

func main() {
	c := &a
	d := &b
	println(*c == *d)
}

// Output:
// true
