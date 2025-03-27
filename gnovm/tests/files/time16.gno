package main

import (
	"time"
)

var t time.Time

func main() {
	println(t.IsDST())
}

// Output:
// false
