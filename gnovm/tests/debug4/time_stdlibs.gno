package main

import "time"

func main() {
	var a int8
	a = 1
	println(time.Second * a)
}

// Error:
// main/debug/time_stdlibs.gno:8: invalid operation: mismatched types time.Duration and int8
