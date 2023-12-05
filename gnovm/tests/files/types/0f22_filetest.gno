package main

var a int = 0
var b int = 1

// TODO: review where it happens
func main() {
	println(&a == &b)
}

// Output:
// false
