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
// files/redeclaration4.gno:8:7: foo redeclared in this block
// 	previous declaration at files/redeclaration4.gno:4:6
