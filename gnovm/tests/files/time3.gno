package main

import (
	"time"
)

// FIXME related to named returns
func main() {
	t := time.Date(2009, time.November, 10, 23, 4, 5, 0, time.UTC)
	println(t.Clock())
}

// Output:
// 23 4 5
