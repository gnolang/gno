package main

var a [2]interface{}
var c [2]interface{}

func gen() interface{} {
	return 1
}

func main() {
	a = [2]interface{}{gen(), gen()}
	c = [2]interface{}{gen(), gen()}
	println(a == c)
}

// Output:
// true
