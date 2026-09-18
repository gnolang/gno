package main

func main() {
	iota := 5 // ERROR "builtin identifiers cannot be shadowed"
	_ = iota
}

// Error:
// main/iota_identifier.gno:4:2-11: builtin identifiers cannot be shadowed: iota
