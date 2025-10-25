// Package pkgerr provides a custom error wrapper that prepends the realm link to your error
// This is useful for identifying the source package of the error.
//
// Usage:
//
// To wrap an error with realm and chain domain information, use the `New` function:
//
//	err := pkgerr.New("my error message") // in gno.land/r/leon/example
//	fmt.Println(err.Error()) // Output: "r/leon/example: my error message"
package pkgerr

import (
	"chain/runtime"
	"errors"
	"strings"

	"gno.land/p/nt/ufmt"
)

// pkgErr is a custom error type that prepends the current
// realm link to the original error message.
type pkgErr struct {
	originalErr error
}

// New creates a new pkgErr with the given error. The returned error will include
// the current realm link in its message.
func New(msg string) error {
	return &pkgErr{originalErr: errors.New(msg)}
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}

	return &pkgErr{originalErr: err}
}

func (e pkgErr) Unwrap() error {
	return e.originalErr
}

// Error implements the `error` interface for pkgErr.
func (e *pkgErr) Error() string {
	return ufmt.Sprintf("%s: %s",
		strings.TrimPrefix(runtime.CurrentRealm().PkgPath(), "gno.land/"),
		e.originalErr)
}
