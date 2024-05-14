package main

var time int

type time string

func main() {
	var t time = "hello"
	println(t)
}

// Error:
// files/redeclaration_global1.gno:5:6: time redeclared in this block
// 	previous declaration at files/redeclaration_global1.gno:3:5
