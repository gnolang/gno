package main

func main() {
	var foo struct {
		yolo string
	}

	var foo int
	foo = 2
	println(foo)
}

// Error:
// files/redeclaration2.gno:8:6: foo redeclared in this block
// 	previous declaration at files/redeclaration2.gno:4:6

// TypeCheckError:
// main/files/redeclaration2.gno:8:6: foo redeclared in this block; main/files/redeclaration2.gno:4:6: 	other declaration of foo; main/files/redeclaration2.gno:9:8: cannot use 2 (untyped int constant) as struct{yolo string} value in assignment
