package main

import "fmt"

func main() {
	const t []string = []string{}
	fmt.Println(t)
}

// Error:
// main/files/const23.gno:6:8: [](const-type string) (variable of type []string) is not constant
