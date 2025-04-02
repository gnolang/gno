package main

import "fmt"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered: ", r)
		}
		panic("other panic")
	}()
	panic("test panic")
}

// Output:
// Recovered:  test panic

// Error:
// panic: test panic[recovered]
// 	panic: other panic
