package main

import "std"

func main() {
	v := std.Coin{}
	const c = v.Denom
}

// Error:
// main/files/const38.gno:7:8: unexpected selector expression type *gnolang.DeclaredType
