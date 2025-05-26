package main

func main() {
	var foo struct {
		yolo string
	}

	type foo struct{}
	var bar foo
	println(bar)
}

// Error:
// redeclaration4.gno:8:7: foo redeclared in this block
// 	previous declaration at redeclaration4.gno:4:6

// TypeCheckError:
// main/redeclaration4.gno:8:7: foo redeclared in this block; main/redeclaration4.gno:4:6: 	other declaration of foo; main/redeclaration4.gno:9:10: foo is not a type
