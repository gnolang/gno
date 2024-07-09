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
	InvalidPkgPathError   struct{ abciError }
	InvalidStmtError      struct{ abciError }
	InvalidExprError      struct{ abciError }
	UnauthorizedUserError struct{ abciError }
	TypeCheckError        struct {
		abciError
		Errors []string `json:"errors"`
	}
)

func (e InvalidPkgPathError) Error() string   { return "invalid package path" }
func (e InvalidStmtError) Error() string      { return "invalid statement" }
func (e InvalidExprError) Error() string      { return "invalid expression" }
func (e UnauthorizedUserError) Error() string { return "unauthorized user" }
func (e TypeCheckError) Error() string {
	var bld strings.Builder
	bld.WriteString("invalid gno package; type check errors:\n")
	bld.WriteString(strings.Join(e.Errors, "\n"))
	return bld.String()
}

func ErrUnauthorizedUser(msg string) error {
	return errors.Wrap(UnauthorizedUserError{}, msg)
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
