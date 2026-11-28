package main

import "chain/banker"

func main() {
	// Type constants are not addressable.
	_ = &banker.BankerTypeReadonly
}

// Error:
// main/addressable_5b_err.gno:7:6-32: invalid operation: cannot take address of (const (0 chain/banker.BankerType)) (value of type chain/banker.BankerType)

// TypeCheckError:
// main/addressable_5b_err.gno:7:7: invalid operation: cannot take address of banker.BankerTypeReadonly (constant 0 of uint8 type banker.BankerType)
