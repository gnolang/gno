package main

import "fmt"

// no support for set `GOEXPERIMENT` for now.
func main() {
	var prints []func()
	for i := 1; i <= 3; i++ {
		v := i
		prints = append(prints, func() { fmt.Println(v) })
	}
	for _, print := range prints {
		print()
	}
}

// Output:
// 3
// 3
// 3

// Go_Output:
// 1
// 2
// 3
