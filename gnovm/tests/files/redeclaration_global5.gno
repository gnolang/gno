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
// redeclaration_global5.gno:5:6: time redeclared in this block
// 	previous declaration at redeclaration_global5.gno:3:5

// TypeCheckError:
// main/redeclaration_global5.gno:5:6: time redeclared in this block; main/redeclaration_global5.gno:3:5: 	other declaration of time; main/redeclaration_global5.gno:10:7: invalid operation: cannot call time (variable of type int): int is not a function
