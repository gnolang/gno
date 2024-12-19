package main

func invalidNative() string

func main() {
	println(invalidNative())
}

// Error:
// main/files/native0.gno:3:1: function invalidNative does not have a body but is not natively defined (did you build after pulling from the repository?)
