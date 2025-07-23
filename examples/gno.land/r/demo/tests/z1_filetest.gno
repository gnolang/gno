package main

import (
	"gno.land/r/demo/tests"
)

func main() {
	println(tests.Counter(cross))
	tests.IncCounter(cross)
	println(tests.Counter(cross))
}

// Output:
// 0
// 1
