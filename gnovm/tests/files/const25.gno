package main

import "fmt"

func main() {
	const t = []string{"1"}
	fmt.Println(t)
}

// Error:
// main/const25.gno:6:8-25: (const-type []string){(const ("1" string))} (variable of type []string) is not constant

// TypeCheckError:
// main/const25.gno:6:12: []string{…} (value of type []string) is not constant
