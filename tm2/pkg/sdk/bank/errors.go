package bank

import (
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// for convenience:
type abciError struct{}

func (abciError) AssertABCIError() {}

// declare all bank errors.
// NOTE: these are meant to be used in conjunction with pkgs/errors.
type NoInputsError struct{ abciError }

type (
	NoOutputsError           struct{ abciError }
	InputOutputMismatchError struct{ abciError }
)

func (e NoInputsError) Error() string  { return "no inputs in send transaction" }
func (e NoOutputsError) Error() string { return "no outputs in send transaction" }
func (e InputOutputMismatchError) Error() string {
	return "sum inputs != sum outputs in send transaction"
}

func ErrNoInputs() error {
	return errors.Wrap(NoInputsError{}, "")
}

func ErrNoOutputs() error {
	return errors.Wrap(NoOutputsError{}, "")
}

func ErrInputOutputMismatch() error {
	return errors.Wrap(InputOutputMismatchError{}, "")
}
