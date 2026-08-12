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
	NoRenderDeclError     struct{ abciError }
	PkgExistError         struct{ abciError }
	InvalidStmtError      struct{ abciError }
	InvalidExprError      struct{ abciError }
	UnauthorizedUserError struct{ abciError }
	InvalidPackageError   struct{ abciError }
	InvalidFileError      struct{ abciError }
	ObjectNotFoundError   struct{ abciError }
	// TypeCheckError deliberately carries no diagnostic strings: it is
	// amino-encoded into ABCIResult.Error, which is merkle-hashed into the
	// block's LastResultsHash, and raw go/types (and go/parser) messages
	// vary in wording, count, and order across Go toolchains — hashing
	// them would let two correctly-rejecting validators commit different
	// result hashes. ErrTypeCheck carries the full messages on the error's
	// msg trace instead, which reaches the user via the unhashed
	// Result.Log.
	TypeCheckError struct{ abciError }
	// UnobservedSendError is returned by MsgCall when a non-empty
	// send-envelope was attached but no executing code ever observed it.
	// Such a call would otherwise strand the coins in the callee's
	// address. See ExecContext.OriginSendObserved.
	UnobservedSendError struct{ abciError }
	// UnspendableSendError is returned by MsgAddPackage when coins are
	// attached to the deployment of a pure `p/` package. Such a package has
	// no realm identity, so it can never obtain a banker and can never
	// spend from its own address — the coins would be lost for good.
	UnspendableSendError struct{ abciError }
)

func (e InvalidPkgPathError) Error() string   { return "invalid package path" }
func (e NoRenderDeclError) Error() string     { return "render function not declared" }
func (e PkgExistError) Error() string         { return "package already exists" }
func (e InvalidStmtError) Error() string      { return "invalid statement" }
func (e InvalidFileError) Error() string      { return "file is not available" }
func (e InvalidExprError) Error() string      { return "invalid expression" }
func (e UnauthorizedUserError) Error() string { return "unauthorized user" }
func (e InvalidPackageError) Error() string   { return "invalid package" }
func (e ObjectNotFoundError) Error() string   { return "object not found" }
func (e TypeCheckError) Error() string        { return "invalid gno package; type check failed" }
func (e UnobservedSendError) Error() string {
	return "coins were sent but the called function never read them"
}

func (e UnspendableSendError) Error() string {
	return "coins cannot be sent to a pure package; nothing could ever spend them"
}

func ErrPkgAlreadyExists(msg string) error {
	return errors.Wrap(PkgExistError{}, msg)
}

func ErrUnauthorizedUser(msg string) error {
	return errors.Wrap(UnauthorizedUserError{}, msg)
}

func ErrInvalidPkgPath(msg string) error {
	return errors.Wrap(InvalidPkgPathError{}, msg)
}

func ErrInvalidFile(msg string) error {
	return errors.Wrap(InvalidFileError{}, msg)
}

func ErrInvalidStmt(msg string) error {
	return errors.Wrap(InvalidStmtError{}, msg)
}

func ErrInvalidExpr(msg string) error {
	return errors.Wrap(InvalidExprError{}, msg)
}

func ErrInvalidPackage(msg string) error {
	return errors.Wrap(InvalidPackageError{}, msg)
}

func ErrObjectNotFound(msg string) error {
	return errors.Wrap(ObjectNotFoundError{}, msg)
}

func ErrUnobservedSend(msg string) error {
	return errors.Wrap(UnobservedSendError{}, msg)
}

func ErrUnspendableSend(msg string) error {
	return errors.Wrap(UnspendableSendError{}, msg)
}

// ErrTypeCheck wraps err's full messages around the empty TypeCheckError
// sentinel. Only the sentinel reaches the hashed tx result; the messages
// ride the msg trace into the unhashed Result.Log (see TypeCheckError).
func ErrTypeCheck(err error) error {
	errs := multierr.Errors(err)
	msgs := make([]string, len(errs))
	for i, err := range errs {
		msgs[i] = err.Error()
	}
	return errors.Wrap(TypeCheckError{}, strings.Join(msgs, "\n"))
}
