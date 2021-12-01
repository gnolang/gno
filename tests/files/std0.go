package main

import "std"

func main() {
	println(std.AddressSize)
	name := "test1"
	if len(name) > std.AddressSize {
		panic("should not happen")
	}
	println("done")
}

// Output:
// 20
// done
