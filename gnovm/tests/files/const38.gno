package main

import "chain"

func main() {
	v := chain.Coin{}
	const c = v.Denom
}

// Error:
// main/const38.gno:7:8-19: v<VPBlock(1,0)>.Denom (variable of type string) is not constant

// TypeCheckError:
// main/const38.gno:7:12: v.Denom (variable of type string) is not constant
