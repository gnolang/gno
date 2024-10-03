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
// files/redeclaration5.gno:8:7: foo redeclared in this block
// 	previous declaration at files/redeclaration5.gno:4:7
