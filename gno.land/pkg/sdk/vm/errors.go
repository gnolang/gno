package vm

import (
	"strings"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"go.uber.org/multierr"
)

// for convenience:
type abciError struct{}

func (abciError) AssertABCIError() {}

// declare all script errors.
// NOTE: these are meant to be used in conjunction with pkgs/errors.
type (
	InvalidPkgPathError struct{ abciError }
	PkgExistError       struct{ abciError }
	InvalidStmtError    struct{ abciError }
	InvalidExprError    struct{ abciError }
	TypeCheckError      struct {
		abciError
		Errors []string
	}
)

func (e InvalidPkgPathError) Error() string { return "invalid package path" }
func (e PkgExistError) Error() string       { return "package already exists" }
func (e InvalidStmtError) Error() string    { return "invalid statement" }
func (e InvalidExprError) Error() string    { return "invalid expression" }
func (e TypeCheckError) Error() string {
	var bld strings.Builder
	bld.WriteString("invalid gno package; type check errors:\n")
	bld.WriteString(strings.Join(e.Errors, "\n"))
	return bld.String()
}

func ErrPkgAlreadyExists(msg string) error {
	return errors.Wrap(PkgExistError{}, msg)
}

func ErrInvalidPkgPath(msg string) error {
	return errors.Wrap(InvalidPkgPathError{}, msg)
}

func ErrInvalidStmt(msg string) error {
	return errors.Wrap(InvalidStmtError{}, msg)
}

func ErrInvalidExpr(msg string) error {
	return errors.Wrap(InvalidExprError{}, msg)
}

func ErrTypeCheck(err error) error {
	var tce TypeCheckError
	errs := multierr.Errors(err)
	for _, err := range errs {
		tce.Errors = append(tce.Errors, err.Error())
	}
	return errors.NewWithData(tce).Stacktrace()
}
