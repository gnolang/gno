package main

var myVar1 = func() {
	a := myDep1
	println(a)
}

var myDep1 string = "hello"

func main() {
	myVar1()
}

// Output:
// hello
