package main

import "fmt"

func main() {
	a := make([]byte, 2147483647)
	fmt.Println(len(a))
}

// Output:
// 2147483647