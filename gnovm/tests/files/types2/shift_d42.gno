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

// TODO: composite literal
// TODO: simple print non-const untyped

func main() {
	type S struct {
		a float32
	}
	s := S{
		a: 1 << 2,
	}
	println(s.a)
}

// Output:
// 4
