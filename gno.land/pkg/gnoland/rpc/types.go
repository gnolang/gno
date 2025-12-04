package rpc

import "github.com/gnolang/gno/tm2/pkg/std"

type SimulateResponse struct {
	GasUsed      int64     `json:"gas_used"`
	StorageFee   std.Coins `json:"storage_fee,omitempty"`
	StorageDelta int64     `json:"storage_delta"` // bytes
}
