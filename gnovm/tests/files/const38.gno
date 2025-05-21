package main

import "std"

func main() {
	v := std.Coin{}
	const c = v.Denom
}

// Error:
// main/files/const38.gno:7:8: v<VPBlock(1,0)>.Denom (variable of type string) is not constant

// TypeCheckError:
// main/files/const38.gno:7:12: v.Denom (variable of type string) is not constant
