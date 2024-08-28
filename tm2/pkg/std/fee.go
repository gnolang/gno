package std

import "github.com/gnolang/gno/tm2/pkg/amino"

// Fee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type Fee struct {
	GasWanted int64 `json:"gas_wanted" yaml:"gas_wanted"`
	GasFee    Coin  `json:"gas_fee" yaml:"gas_fee"`
}

// NewFee returns a new instance of Fee
func NewFee(gasWanted int64, gasFee Coin) Fee {
	return Fee{
		GasWanted: gasWanted,
		GasFee:    gasFee,
	}
}

// Bytes for signing later
func (fee Fee) Bytes() []byte {
	bz, err := amino.MarshalJSON(fee) // TODO
	if err != nil {
		panic(err)
	}
	return bz
}
