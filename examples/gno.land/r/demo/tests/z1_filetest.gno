package main

import (
	"gno.land/r/demo/tests"
)

func main() {
	println(tests.Counter())
	tests.IncCounter()
	println(tests.Counter())
}

// Output:
// 0
// 1
