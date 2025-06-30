package main

func main() {
	var foo string

	var foo int
	foo = 2
	println(foo)
}

// Error:
// redeclaration1.gno:6:6: foo redeclared in this block
// 	previous declaration at redeclaration1.gno:4:6

// TypeCheckError:
// main/redeclaration1.gno:6:6: foo redeclared in this block; main/redeclaration1.gno:4:6: 	other declaration of foo; main/redeclaration1.gno:7:8: cannot use 2 (untyped int constant) as string value in assignment
