package main

import "time"

type Duration struct {
	t time.Duration
}

func main() {
	var a, b Duration
	println(a == b)
}

// Output:
// true
