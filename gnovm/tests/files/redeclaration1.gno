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

// TypeCheckError:
// main/files/redeclaration1.gno:6:6: foo redeclared in this block; main/files/redeclaration1.gno:4:6: 	other declaration of foo; main/files/redeclaration1.gno:7:8: cannot use 2 (untyped int constant) as string value in assignment
