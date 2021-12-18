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
// files2/redeclaration-global5.go:5:6: time redeclared in this block
//	previous declaration at files2/redeclaration-global5.go:3:5
