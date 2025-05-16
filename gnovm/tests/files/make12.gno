package main

import "fmt"

func main() {
	a := make([]int, 32767)
	fmt.Println(len(a))
}

// Output:
// 32767
