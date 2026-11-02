package main

func main() {
	uint256 := 3
	println(uint256)
}

// Error:
// main/err_shadow_local.gno:4:2-9: builtin identifiers cannot be shadowed: uint256
