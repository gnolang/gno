package main

func main() {
	type foo struct {
		yolo string
	}

	var foo int
	foo = 2
	println(foo)
}

// Error:
// files/redeclaration0.gno:8:6: foo redeclared in this block
// 	previous declaration at files/redeclaration0.gno:4:7
