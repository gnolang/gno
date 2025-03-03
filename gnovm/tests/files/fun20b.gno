package main

import "fmt"

var myerr any = fmt.Errorf("bar")

func ferr() any { return myerr }

func foo() ([]string, any) {
	return nil, ferr()
}

func main() {
	a, b := foo()
	fmt.Println(a, b)
}

// Output:
// [] bar
