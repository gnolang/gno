package main

// int256 and uint256 are builtin identifiers, which Gno does not let any
// declaration shadow, imports included: the library package
// gno.land/p/onbloc/uint256 must be imported under another name.
import "gno.land/p/onbloc/uint256/v0"

func main() {
	println(uint256.NewUint(1).Dec())
}

// Error:
// main/err_shadow_import.gno:6:8-38: builtin identifiers cannot be shadowed: uint256
