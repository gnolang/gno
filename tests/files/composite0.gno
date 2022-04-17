package main

import "fmt"

var a = &[]*T{}

// NOTE:
// lowercase fields don't work with native funcs.
// see tests/files/composite0b.go for gno println.
type T struct{ Name string }

func main() {
	fmt.Println(a)
}

// Output:
// &[]
