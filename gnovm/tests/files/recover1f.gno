package main

import "fmt"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered: ", r)
		}
	}()

	defer func() {
		panic("second")
	}()

	panic("first")
}

// Output:
// recovered:  second
