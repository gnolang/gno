package vm

import "github.com/gnolang/gno/pkgs/errors"

// for convenience:
type abciError struct{}

func (_ abciError) AssertABCIError() {}

// declare all script errors.
// NOTE: these are meant to be used in conjunction with pkgs/errors.
type InvalidPkgPathError struct{ abciError }
type InvalidStmtError struct{ abciError }
type InvalidExprError struct{ abciError }

func (e InvalidPkgPathError) Error() string { return "invalid package path" }
func (e InvalidStmtError) Error() string    { return "invalid statement" }
func (e InvalidExprError) Error() string    { return "invalid expression" }

func ErrInvalidPkgPath(msg string) error {
	return errors.Wrap(InvalidPkgPathError{}, msg)
}

func ErrInvalidStmt(msg string) error {
	return errors.Wrap(InvalidStmtError{}, msg)
}

func ErrInvalidExpr(msg string) error {
	return errors.Wrap(InvalidExprError{}, msg)
}
