package main

func f(iota int) { // ERROR "builtin identifiers cannot be shadowed"
	println(iota)
}

func main() {
	f(3)
}

// Error:
// main/iota_identifier_param.gno:3:1-5:2: builtin identifiers cannot be shadowed: iota

