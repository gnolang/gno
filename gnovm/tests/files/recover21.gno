package main

import "fmt"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered: ", r)
		}
	}()

	v := -1
	_ = 1 << v
}

// Output:
// recovered:  runtime error: negative shift amount: (-1 int)
