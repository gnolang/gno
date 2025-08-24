package main

import "fmt"

func main() {
	const t []string = []string{}
	fmt.Println(t)
}

// Error:
// main/const23.gno:6:8-31: invalid constant type []string

// TypeCheckError:
// main/const23.gno:6:10: invalid constant type []string
