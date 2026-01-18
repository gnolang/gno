package main

import "time"

func main() {
	println(time.UTC == nil)
	time.UTC = nil
	println(time.UTC == nil)
	println("done")
}

// Output:

// Error:
// main/stdlibs_native.gno:7:2-10: cannot directly mutate time.UTC from main
