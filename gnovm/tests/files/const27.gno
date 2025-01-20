package main

import "fmt"

func v() string {
	return ""
}

func main() {
	var i interface{} = 1
	const t, ok = i.(int)
	fmt.Println(t, ok)
}

// Error:
// main/files/const27.gno:11:8: i<VPBlock(1,0)>.((const-type int)) (comma, ok expression of type (const-type int)) is not constant
