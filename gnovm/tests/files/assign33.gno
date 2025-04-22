package main

func foo() (int, bool) {
	return 1, true
}

func main() {
	var a, b int = foo()
	println(a, b)
}

// Error:
// main/files/assign33.gno:8:6: cannot use foo<VPBlock(3,0)>() (value of type bool) as int value in assignment

// TypeCheckError:
// main/files/assign33.gno:8:17: cannot use foo() (value of type bool) as int value in assignment
