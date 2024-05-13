package main

func main() {}

var myVar = func() { hello(myDep) }

var myDep string

var myVar1 = func() {
	a := myDep1
}

var myDep1 string

// Output:
// 123
