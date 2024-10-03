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
// true
// done
