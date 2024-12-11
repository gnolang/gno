package main

import "fmt"

func main() {
	const t = []string{"1"}
	fmt.Println(t)
}

// Error:
// main/files/const25.gno:6:8: [](const-type string) (variable of type []string) is not constant
