package main

import "chain/banker"

func main() {
	v := banker.Coin{}
	const c = v.Denom
}

// Error:
// main/files/const38.gno:7:8: v<VPBlock(1,0)>.Denom (variable of type string) is not constant
