package main

import "std"

func main() {
	defer func() {
		// assert panic is recoverable
		println(recover())
	}()
	std.VerifySignature("","","")
}

// Error:
// decoding Bech32 failed: must provide a valid bech32 string
