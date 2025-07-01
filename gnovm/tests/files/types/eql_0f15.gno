package main

var a [2]any
var c [2]any

func gen() any {
	return 1
}

func main() {
	a = [2]any{gen(), gen()}
	c = [2]any{gen(), gen()}
	println(a == c)
}

// Output:
// true
