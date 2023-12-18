package main

var a [2]string
var b [2]string

func main() {
	a = [2]string{"hello", "world"}
	b = [2]string{"hello", "world"}
	println(a == b)
}

// Output:
// true
