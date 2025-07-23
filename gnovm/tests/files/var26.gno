package main

import "fmt"

func main() {
	var a, c = 1, 2, 3
	fmt.Println(a, c)
}

// Error:
// main/var26.gno:6:6-20: extra init expr 3

// TypeCheckError:
// main/var26.gno:6:19: extra init expr 3
