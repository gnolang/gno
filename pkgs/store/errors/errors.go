package errors

import (
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
)

// nolint - reexport
type Error = abci.Error

// nolint - reexport
func ErrInternal(msg string) Error {
	return abci.StringError("internal:" + msg)
}
func ErrTxDecode(msg string) Error {
	return abci.StringError("txdecode:" + msg)
}
func ErrUnknownRequest(msg string) Error {
	return abci.StringError("unknownrequest:" + msg)
}
