package main

func main() {
	myDep = "123"
	myVar()
}

func hello(s string) {}

var myVar = func() { hello(myDep) }

var myDep string

// Output:
// 123
