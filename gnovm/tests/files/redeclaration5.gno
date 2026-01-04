package main

func main() {
	type foo struct {
		yolo string
	}

	type foo struct{}
	var bar foo
	println(bar)
}

// Error:
// redeclaration5.gno:8:7: foo redeclared in this block
// 	previous declaration at redeclaration5.gno:4:7

// TypeCheckError:
// main/redeclaration5.gno:8:7: foo redeclared in this block; main/redeclaration5.gno:4:7: 	other declaration of foo
