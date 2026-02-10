// Package genesis provides a way to store and access dynamic variables that are
// set during the genesis or block0 initialization phase. This package
// demonstrates an important aspect of p/ packages in Gno: while they are
// pure packages, they can have mutable state during initialization (init phase)
// before becoming read-only.
//
// Future improvements:
// When Gno supports something similar to "go build -X", this package could be
// enhanced to accept variables from the CLI or environment variables (e.g., in
// gnodev context). For now, it only supports initialization-time operations
// that make sense at the beginning of the system's lifecycle.
package genesis

import (
	"chain/runtime"
	"time"
)

var (
	// Time is the time of the genesis block.
	Time = time.Now()

	// Height is the height of the genesis block (usually 0).
	Height = runtime.ChainHeight()

	// Domain is the domain of the chain.
	Domain = runtime.ChainDomain()

	// XXX: TZ
	// XXX: Supply = std.Coins{{"ugnot", std.NewBanker(std.BankerTypeReadonly).TotalSupply("ugnot")}}
)

// Uptime returns the uptime of the chain.
func Uptime() time.Duration {
	return time.Since(Time)
}

// Upheight returns the height of the chain.
func Upheight() int64 {
	return runtime.ChainHeight() - Height
}
