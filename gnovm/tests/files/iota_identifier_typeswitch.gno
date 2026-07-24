package main

func main() {
	var x interface{} = 5
	switch iota := x.(type) { // ERROR "builtin identifiers cannot be shadowed"
	case int:
		println(iota)
	}
}

// Error:
// main/iota_identifier_typeswitch.gno:6:2-7:16: builtin identifiers cannot be shadowed: iota

