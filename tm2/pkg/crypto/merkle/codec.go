package merkle

import (
	amino "github.com/gnolang/gno/tm2/pkg/amino"
)

var cdc *amino.Codec

func init() {
	cdc = amino.NewCodec()
	cdc.Seal()
}
