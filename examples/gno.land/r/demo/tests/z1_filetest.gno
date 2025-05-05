package main

import (
	"gno.land/r/demo/tests"
)

func main() {
	println(cross(tests.Counter)())
	cross(tests.IncCounter)()
	println(cross(tests.Counter)())
}

// Output:
// 0
// 1
