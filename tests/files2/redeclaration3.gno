package main

func main() {
	var foo int
	foo = 2

	type foo struct{}
	var bar foo
	println(bar)
}

// Error:
// files2/redeclaration3.gno:7:7: foo redeclared in this block
// 	previous declaration at files2/redeclaration3.gno:4:6