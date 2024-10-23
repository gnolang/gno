package main

import (
	"time"
)

func main() {
	year, month, day := time.Now().Date()
	println("year", year, "month", month, "day", day)
	it := time.Now().InternalDebug()
	println("InternalDebug", it)
}

// Output:
// 2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
