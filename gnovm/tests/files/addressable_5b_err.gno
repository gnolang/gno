package main

import "std"

func main() {
	// Type constants are not addressable.
	_ = &std.BankerTypeReadonly
}

// Error:
// main/files/addressable_5b_err.gno:7:6: cannot take address of (const (0 std.BankerType))
