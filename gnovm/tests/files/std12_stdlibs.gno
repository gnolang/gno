package main

import "std"

func main() {
	std.SetParamString("foo.string", "hello")
	std.SetParamInt64("bar.int64", 12345)
	std.SetParamInt64("baz.uint64", -12345)
	std.SetParamInt64("oof.bool", true)
	std.SetParamInt64("rab.bytes", []byte("world"))
	println("done")
}

// Output:
// done
