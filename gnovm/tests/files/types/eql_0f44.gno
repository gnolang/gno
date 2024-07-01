package main

import "fmt"

type f func()

func main() {
	if f(nil) == nil { // rcx.T == nil
		fmt.Println("func is nil")
	} else {
		fmt.Println("func is not nil")
	}
}

// Output:
// func is nil
