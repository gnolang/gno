package main

func main() {
	var foo int
	foo = 2

	type foo struct{}
	var bar foo
	println(bar)
}

// Error:
// redeclaration3.gno:7:7: foo redeclared in this block
// 	previous declaration at redeclaration3.gno:4:6

// TypeCheckError:
// main/redeclaration3.gno:7:7: foo redeclared in this block; main/redeclaration3.gno:4:6: 	other declaration of foo; main/redeclaration3.gno:8:10: foo is not a type
