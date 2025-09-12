package main

import (
	tests "gno.land/r/tests/vm"
)

func main() {
	println(tests.Counter(cross))
	tests.IncCounter(cross)
	println(tests.Counter(cross))
}

// Output:
// 0
// 1
