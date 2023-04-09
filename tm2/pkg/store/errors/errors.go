package errors

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

type Error = abci.Error

// XXX DEPRECATE AND USE STD.ERRORS
func ErrInternal(msg string) Error {
	return abci.StringError("internal error:" + msg)
}

func ErrTxDecode(msg string) Error {
	return abci.StringError("txdecode error:" + msg)
}

func ErrUnknownRequest(msg string) Error {
	return abci.StringError("unknownrequest error:" + msg)
}
