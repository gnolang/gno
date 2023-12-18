package main

import "fmt"

func main() {
	if func() {} == nil {
		fmt.Println("func is nil")
	} else {
		fmt.Println("func is not nil")
	}
}

// Output:
// func is not nil
