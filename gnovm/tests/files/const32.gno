package main

import "fmt"

func main() {
	const t = 1 + 2 + len([]string{})
	fmt.Println(t)
}

// Error:
// main/const32.gno:6:8-35: (const-type []string){} (variable of type []string) is not constant

// TypeCheckError:
// main/const32.gno:6:12: 1 + 2 + len([]string{}) (value of type int) is not constant
