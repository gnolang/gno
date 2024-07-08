package main

import (
	ptests "gno.land/p/demo/tests"
	rtests "gno.land/r/demo/tests"
)

func main() {
	println(rtests.Counter())
	ptests.IncCounter()
	println(rtests.Counter())
}

// Output:
// 0
// 1
