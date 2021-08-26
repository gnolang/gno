package errors

import (
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
)

// nolint - reexport
type Error = abci.Error

// nolint - reexport
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
