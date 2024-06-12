package main

import "fmt"

func main() {
	const Huge = 1e1000
	fmt.Println(Huge / 1e999)
}

// Output:
// 10
