package merkle

import (
	amino "github.com/gnolang/gno/pkgs/amino"
)

var cdc *amino.Codec

func init() {
	cdc = amino.NewCodec()
	cdc.Seal()
}
