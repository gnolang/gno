package main

import (
	"time"
)

func main() {
	println((5 * time.Minute).Seconds())
}

// Output:
// 300
