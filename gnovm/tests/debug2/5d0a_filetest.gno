package main

// both untyped const
// TODO: dec value representation, and this should happen in process stage!!!
func main() {
	println(1.0 % 1)
}

// Error:
// main/debug/5d0a_filetest.gno:6: operator % not defined on: BigdecKind