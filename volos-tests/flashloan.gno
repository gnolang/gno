package main

import (
	"gno.land/r/volos_mocks"
)

func main() {
	// Test flash loan with GNS token
	volos_mocks.FlashLoan(cross, "gno.land/r/gnoswap/v1/gns", 10000, []byte("gno.land/r/gnoswap/v1/gns"))
} 