package main

import "fmt"

func doRecover2() {
	if r := recover(); r != nil {
		fmt.Println("recovered: ", r)
	} else {
		fmt.Println("not recovered, r: ", r)
	}
}

func doRecover() {
	doRecover2()
}

func main() {
	defer func() {
		doRecover()
	}()
	panic("test panic")
}

// Error:
// panic: test panic

// Output:
// not recovered, r:  <nil>
