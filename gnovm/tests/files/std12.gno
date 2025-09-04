package main

import "std"

func main() {
	std.SetParamString("foo.string", "hello")
	std.SetParamInt64("bar.int64", -12345)
	std.SetParamUint64("baz.uint64", 12345)
	std.SetParamBool("oof.bool", true)
	std.SetParamBytes("rab.bytes", []byte("world"))
	println("done")
}

// Output:
// done
