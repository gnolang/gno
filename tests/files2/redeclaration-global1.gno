package main

var time int

type time string

func main() {
	var t time = "hello"
	println(t)
}

// Error:
// files2/redeclaration-global1.gno:5:6: time redeclared in this block
// 	previous declaration at files2/redeclaration-global1.gno:3:5