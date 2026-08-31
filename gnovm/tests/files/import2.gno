package main

import . "strings"

func main() {
	println(HasPrefix("ab", "a"))
}

// Error:
// main/import2.gno:3:8-19: dot imports not allowed in gno
