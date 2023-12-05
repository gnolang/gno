package main

import "fmt"

func main() {
	x := "hello"
	y := +x
	fmt.Println(y)
}

// Error:
// main/files/types/23a0c_filetest.gno:7: operator + not defined on: string
