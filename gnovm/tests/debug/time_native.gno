package main

import "time"

func main() {
	var a int8
	a = 1
	println(time.Second * a)
}

// Error:
// main/debug/time_native.gno:8: cannot use int8 as int64
