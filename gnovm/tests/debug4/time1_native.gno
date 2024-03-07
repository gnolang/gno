package main

import "time"

func main() {
	var a int8
	a = 1
	println(a * time.Second)
}

// Error:
// main/debug/time1_native.gno:8: cannot use int8 as int64
