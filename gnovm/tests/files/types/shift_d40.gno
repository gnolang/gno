package main

// 1. find where shift expr is used
// a) type cast
// b) assign stmt
// c) composite literals

// 2. get the dest type, if convert condition, convert
// condition:
// a) arg has shift and untyped. since everytime we deal with non-composed shift expr, we can set and get
//    attr: HAS_SHIFT to determine if has shift.

// 3. in type func, find termination condition when not const
//    this condition should be the opposite of the convert condition

// TODO: composite literal, checked, OK!
// TODO: simple print non-const untyped, checked, to typed if interface, OK!

func main() {
	a := 2
	s := []float32{1 << a}
	println(s[0])
}

// Error:
// main/files/types/shift_d40.gno:21:7: operator << not defined on: Float32Kind
