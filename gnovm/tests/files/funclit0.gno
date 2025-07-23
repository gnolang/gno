package main

// This tests an issue where the preprocessor was failing after
// implementing ATTR_PREPROCESS_SKIPPED.

func main() {
	var myDep1 string = "hello"

	var myVar1 = func() {
		a := myDep1
		println(a)
	}

	myVar1()
}

// Output:
// hello
