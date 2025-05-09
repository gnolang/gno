package main

var time int

func time() string {
	return "hello"
}

func main() {
	t := time()
	println(t)
}

// Error:
// files/redeclaration_global5.gno:5:6: time redeclared in this block
// 	previous declaration at files/redeclaration_global5.gno:3:5

// TypeCheckError:
// main/files/redeclaration_global5.gno:5:6: time redeclared in this block; main/files/redeclaration_global5.gno:3:5: 	other declaration of time; main/files/redeclaration_global5.gno:10:7: invalid operation: cannot call non-function time (variable of type int)
