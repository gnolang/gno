package main

// This tests an issue where the preprocessor was failing after
// implementing ATTR_PREPROCESS_SKIPPED.

var myVar0 = func() {
	var myDep1 string = "hello"

	var myVar1 = func() {
		a := myDep1
		println(a)
	}
	myVar1()

	println(myDep0)
}

var myDep0 string = "world"

func main() {

	myVar0()
}

// Output:
// hello
// world
