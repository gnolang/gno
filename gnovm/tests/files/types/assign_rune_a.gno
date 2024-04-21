package main

var r rune

// assign
func main() {
	r = "a"
	println(r)
}

// Error:
// main/files/types/assign_rune_a.gno:7: cannot use untyped string as Int32Kind
