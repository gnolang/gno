package main

import "std"

func main() {
	println(std.RawAddressSize)
	name := "test1"
	if len(name) > std.RawAddressSize {
		panic("should not happen")
	}
	println("done")
}

// Output:
// 20
// done
