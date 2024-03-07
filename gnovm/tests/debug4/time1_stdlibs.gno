package main

import "time"

func main() {
	var a int8
	a = 1
	println(a * time.Second)
}

// Error:
// main/debug/time1_stdlibs.gno:8: invalid operation: mismatched types int8 and time.Duration
