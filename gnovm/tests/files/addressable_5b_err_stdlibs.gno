package main

import "chain/banker"

func main() {
	// Type constants are not addressable.
	_ = &banker.BankerTypeReadonly
}

// Error:
// main/files/addressable_5b_err_stdlibs.gno:7:6: cannot take address of (const (0 chain/banker.BankerType))
