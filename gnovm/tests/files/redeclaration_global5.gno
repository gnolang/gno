package main

var time int

func time() string {
	return "hello"
}

func main() {
	t := time()
	println(t)
}

// Error:
// files/redeclaration_global5.gno:5:6: time redeclared in this block
// 	previous declaration at files/redeclaration_global5.gno:3:5
