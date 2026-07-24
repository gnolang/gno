package main

type T int

func (iota T) M() { // ERROR "builtin identifiers cannot be shadowed"
	println(iota)
}

func main() {
	var t T
	t.M()
}

// Error:
// main/iota_identifier_recv.gno:5:1-7:2: builtin identifiers cannot be shadowed: iota

