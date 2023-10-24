package std

import (
	"github.com/gnolang/gno/tm2/pkg/amino"

	_ "github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/mock"
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
