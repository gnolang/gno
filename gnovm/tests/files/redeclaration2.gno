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
