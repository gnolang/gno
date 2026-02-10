package main

import "chain/banker"

func main() {
	// Type constants are not addressable.
	_ = &banker.BankerTypeReadonly
}

// Error:
// illegal assignment X expression type *gnolang.ConstExpr

// TypeCheckError:
// main/addressable_5b_err.gno:7:7: invalid operation: cannot take address of banker.BankerTypeReadonly (constant 0 of uint8 type banker.BankerType)
