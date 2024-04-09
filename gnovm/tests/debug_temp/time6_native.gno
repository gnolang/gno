package main

import (
	"time"
)

func main() {
	t := &time.Time{}
	println(t.Second() + t.Second())
}

// Output:
