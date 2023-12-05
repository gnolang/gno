package main

import "fmt"

type MyFunc func()

func main() {
	var f *MyFunc

	fmt.Printf("%T \n", f)

	println(f)
}

// Error:
// not yet supported