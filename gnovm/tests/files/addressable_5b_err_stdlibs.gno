package main

import "std"

func main() {
	// Type constants are not addressable.
	_ = &std.BankerTypeReadonly
}

// Error:
// main/files/addressable_5b_err_stdlibs.gno:7:6: illegal assignment X expression type *gnolang.ConstExpr
