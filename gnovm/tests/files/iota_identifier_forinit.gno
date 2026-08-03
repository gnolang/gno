package main

func main() {
	for iota := 0; iota < 2; iota++ { // ERROR "builtin identifiers cannot be shadowed"
		println(iota)
	}
}

// Error:
// main/iota_identifier_forinit.gno:4:2-6:3: builtin identifiers cannot be shadowed: iota

