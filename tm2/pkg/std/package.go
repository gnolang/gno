package std

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/std",
	"std",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(

	// Account
	&BaseAccount{}, "BaseAccount",

	// MemFile/MemPackage
	MemFile{}, "MemFile",
	MemPackage{}, "MemPackage",

	// Errors
	InternalError{}, "InternalError",
	TxDecodeError{}, "TxDecodeError",
	InvalidSequenceError{}, "InvalidSequenceError",
	UnauthorizedError{}, "UnauthorizedError",
	InsufficientFundsError{}, "InsufficientFundsError",
	UnknownRequestError{}, "UnknownRequestError",
	InvalidAddressError{}, "InvalidAddressError",
	UnknownAddressError{}, "UnknownAddressError",
	InvalidPubKeyError{}, "InvalidPubKeyError",
	InsufficientCoinsError{}, "InsufficientCoinsError",
	InvalidCoinsError{}, "InvalidCoinsError",
	InvalidGasWantedError{}, "InvalidGasWantedError",
	OutOfGasError{}, "OutOfGasError",
	MemoTooLargeError{}, "MemoTooLargeError",
	InsufficientFeeError{}, "InsufficientFeeError",
	TooManySignaturesError{}, "TooManySignaturesError",
	NoSignaturesError{}, "NoSignaturesError",
	GasOverflowError{}, "GasOverflowError",
))
