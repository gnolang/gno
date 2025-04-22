package main

type time int

var time string

func main() {
	time = "hello"
	println(time)
}

// Error:
// files/redeclaration_global0.gno:5:5: time redeclared in this block
// 	previous declaration at files/redeclaration_global0.gno:3:6

// TypeCheckError:
// main/files/redeclaration_global0.gno:5:5: time redeclared in this block; main/files/redeclaration_global0.gno:3:6: 	other declaration of time; main/files/redeclaration_global0.gno:8:2: time (type) is not an expression; main/files/redeclaration_global0.gno:9:10: time (type) is not an expression
