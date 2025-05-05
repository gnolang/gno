package main

import "time"

func main() {
	println(time.UTC == nil)
	time.UTC = nil
	println(time.UTC == nil)
	println("done")
}

// Output:
// false

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(time) package{})).UTC
