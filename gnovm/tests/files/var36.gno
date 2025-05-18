package main

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
