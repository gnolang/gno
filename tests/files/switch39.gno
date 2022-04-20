package main

import "fmt"

func main() {
	switch true {
	case false:
		var x = "dontcare"
		fmt.Println(x)
		panic("should not happen")
	case true:
		var x = "apples"
		var y = "oranges"
		fmt.Println(x, y)
	default:
		panic("strange case but ok")
	}
}

// Output:
// apples oranges
