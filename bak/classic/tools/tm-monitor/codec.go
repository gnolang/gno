package main

import (
	amino "github.com/tendermint/go-amino-x"
	ctypes "github.com/tendermint/classic/rpc/core/types"
)

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}
