package main

import "std"

func main() {
	// Type constants are not addressable.
	_ = &std.BankerTypeReadonly
}

// Error:
// illegal assignment X expression type *gnolang.ConstExpr

// TypeCheckError:
// main/addressable_5b_err_stdlibs.gno:7:7: invalid operation: cannot take address of std.BankerTypeReadonly (constant 0 of uint8 type std.BankerType)
