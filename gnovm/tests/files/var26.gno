package main

import "fmt"

func main() {
	var a, c = 1, 2, 3
	fmt.Println(a, c)
}

// Error:
// main/files/var26.gno:6:6: extra init expr 3
