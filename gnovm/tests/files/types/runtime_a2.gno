package main

import "fmt"

func gen() interface{} {
	return false
}

func main() {
	r := gen()
	fmt.Printf("%T \n", r)
}

// Output:
// bool
