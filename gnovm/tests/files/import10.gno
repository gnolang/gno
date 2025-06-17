package main

import "fmt"

func main() {
	fmt.Println(1)
	println(fmt)
}

// Error:
// main/import10.gno:7:10-13: package fmt cannot only be referred to in a selector expression

// TypeCheckError:
// main/import10.gno:7:10: use of package fmt not in selector
