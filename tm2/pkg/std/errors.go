package std

import (
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// for convenience:
type abciError struct{}

func (abciError) AssertABCIError() {}

// declare all std errors.
// NOTE: these are meant to be used in conjunction with pkgs/errors.
type InternalError struct{ abciError }

type (
	TxDecodeError          struct{ abciError }
	InvalidSequenceError   struct{ abciError }
	UnauthorizedError      struct{ abciError }
	InsufficientFundsError struct{ abciError }
	UnknownRequestError    struct{ abciError }
	InvalidAddressError    struct{ abciError }
	UnknownAddressError    struct{ abciError }
	InvalidPubKeyError     struct{ abciError }
	InsufficientCoinsError struct{ abciError }
	InvalidCoinsError      struct{ abciError }
	InvalidGasWantedError  struct{ abciError }
	OutOfGasError          struct{ abciError }
	MemoTooLargeError      struct{ abciError }
	InsufficientFeeError   struct{ abciError }
	TooManySignaturesError struct{ abciError }
	NoSignaturesError      struct{ abciError }
	GasOverflowError       struct{ abciError }
)

func (e InternalError) Error() string          { return "internal error" }
func (e TxDecodeError) Error() string          { return "tx decode error" }
func (e InvalidSequenceError) Error() string   { return "invalid sequence error" }
func (e UnauthorizedError) Error() string      { return "unauthorized error" }
func (e InsufficientFundsError) Error() string { return "insufficient funds error" }
func (e UnknownRequestError) Error() string    { return "unknown request error" }
func (e InvalidAddressError) Error() string    { return "invalid address error" }
func (e UnknownAddressError) Error() string    { return "unknown address error" }
func (e InvalidPubKeyError) Error() string     { return "invalid pubkey error" }
func (e InsufficientCoinsError) Error() string { return "insufficient coins error" }
func (e InvalidCoinsError) Error() string      { return "invalid coins error" }
func (e InvalidGasWantedError) Error() string  { return "invalid gas wanted" }
func (e OutOfGasError) Error() string          { return "out of gas error" }
func (e MemoTooLargeError) Error() string      { return "memo too large error" }
func (e InsufficientFeeError) Error() string   { return "insufficient fee error" }
func (e TooManySignaturesError) Error() string { return "too many signatures error" }
func (e NoSignaturesError) Error() string      { return "no signatures error" }
func (e GasOverflowError) Error() string       { return "gas overflow error" }

// NOTE also update pkg/std/package.go registrations.

func ErrInternal(msg string) error {
	return errors.Wrap(InternalError{}, msg)
}

func ErrTxDecode(msg string) error {
	return errors.Wrap(TxDecodeError{}, msg)
}

func ErrInvalidSequence(msg string) error {
	return errors.Wrap(InvalidSequenceError{}, msg)
}

func ErrUnauthorized(msg string) error {
	return errors.Wrap(UnauthorizedError{}, msg)
}

func ErrInsufficientFunds(msg string) error {
	return errors.Wrap(InsufficientFundsError{}, msg)
}

func ErrUnknownRequest(msg string) error {
	return errors.Wrap(UnknownRequestError{}, msg)
}

func ErrInvalidAddress(msg string) error {
	return errors.Wrap(InvalidAddressError{}, msg)
}

func ErrUnknownAddress(msg string) error {
	return errors.Wrap(UnknownAddressError{}, msg)
}

func ErrInvalidPubKey(msg string) error {
	return errors.Wrap(InvalidPubKeyError{}, msg)
}

func ErrInsufficientCoins(msg string) error {
	return errors.Wrap(InsufficientCoinsError{}, msg)
}

func ErrInvalidCoins(msg string) error {
	return errors.Wrap(InvalidCoinsError{}, msg)
}

func ErrInvalidGasWanted(msg string) error {
	return errors.Wrap(InvalidGasWantedError{}, msg)
}

func ErrOutOfGas(msg string) error {
	return errors.Wrap(OutOfGasError{}, msg)
}

func ErrMemoTooLarge(msg string) error {
	return errors.Wrap(MemoTooLargeError{}, msg)
}

func ErrInsufficientFee(msg string) error {
	return errors.Wrap(InsufficientFeeError{}, msg)
}

func ErrTooManySignatures(msg string) error {
	return errors.Wrap(TooManySignaturesError{}, msg)
}

func ErrNoSignatures(msg string) error {
	return errors.Wrap(NoSignaturesError{}, msg)
}

func ErrGasOverflow(msg string) error {
	return errors.Wrap(GasOverflowError{}, msg)
}
