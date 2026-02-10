package main

import "chain/params"

func main() {
	params.SetString("foo.string", "hello")
	params.SetInt64("bar.int64", -12345)
	params.SetUint64("baz.uint64", 12345)
	params.SetBool("oof.bool", true)
	params.SetBytes("rab.bytes", []byte("world"))
	println("done")
}

// Output:
// done
