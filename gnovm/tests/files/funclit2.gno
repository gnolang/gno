package main

// This tests an issue where the preprocessor was failing after
// implementing ATTR_PREPROCESS_SKIPPED.

var myVar0 = func() {

	var myVar1 = func() {
		a := myDep1
		println(a)
	}

	// invalid
	var myDep1 string = "hello"

	myVar1()

	println(myDep0)
}

var myDep0 string = "world"

func main() {

	myVar0()
}

// Error:
// main/funclit2.gno:9:8-14: name myDep1 not declared

// TypeCheckError:
// main/funclit2.gno:9:8: undefined: myDep1; main/funclit2.gno:14:6: declared and not used: myDep1
