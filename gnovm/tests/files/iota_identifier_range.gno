package main

func main() {
	s := []int{1, 2, 3}
	for iota := range s { // ERROR "builtin identifiers cannot be shadowed"
		_ = iota
	}
}

// Error:
// main/iota_identifier_range.gno:5:2-7:3: builtin identifiers cannot be shadowed: iota
