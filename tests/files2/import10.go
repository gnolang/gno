package main

import "fmt"

func main() {
	fmt.Println(1)
	println(fmt)
}

// Error:
// package fmt cannot only be referred to in a selector expression
