package main

import "fmt"

func main() {
	fmt.Println(1)
	println(fmt)
}

// Error:
// main/files/import10.gno:7:10: package fmt cannot only be referred to in a selector expression
