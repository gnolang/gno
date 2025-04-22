package main

var r rune

// assign
func main() {
	r = "a"
	println(r)
}

// Error:
// main/files/types/assign_rune_a.gno:7:2: cannot use untyped string as Int32Kind

// TypeCheckError:
// main/files/types/assign_rune_a.gno:7:6: cannot use "a" (untyped string constant) as rune value in assignment
