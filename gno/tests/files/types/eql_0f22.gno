package main

var a int = 0
var b int = 1

func main() {
	println(&a == &b)
}

// Output:
// false
