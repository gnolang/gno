package main

import "strconv"

var value int = 0

func Foo(_ string) string { return strconv.Itoa(value) }

func main() {
	println(Foo(""))
}

// Output:
// 0
