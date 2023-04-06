package main

import "fmt"

var myerr interface{} = fmt.Errorf("bar")

func ferr() interface{} { return myerr }

func foo() ([]string, interface{}) {
	return nil, ferr()
}

func main() {
	a, b := foo()
	fmt.Println(a, b)
}

// Output:
// [] bar
