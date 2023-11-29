package main

func main() {
	var foo string

	var foo int
	foo = 2
	println(foo)
}

// Error:
// files/redeclaration1.gno:6:6: foo redeclared in this block
// 	previous declaration at files/redeclaration1.gno:4:6
