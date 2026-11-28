package main

func f() (iota int) { // ERROR "builtin identifiers cannot be shadowed"
	return 5
}

func main() {
	println(f())
}

// Error:
// main/iota_identifier_result.gno:3:1-5:2: builtin identifiers cannot be shadowed: iota

// TypeCheckError:
// main/iota_identifier_result.gno:3:11: builtin identifiers cannot be shadowed: iota
