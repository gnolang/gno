package main

var time int

type time string

func main() {
	var t time = "hello"
	println(t)
}

// Error:
// redeclaration_global1.gno:5:6: time redeclared in this block
// 	previous declaration at redeclaration_global1.gno:3:5

// TypeCheckError:
// main/redeclaration_global1.gno:5:6: time redeclared in this block; main/redeclaration_global1.gno:3:5: 	other declaration of time; main/redeclaration_global1.gno:8:8: time is not a type
