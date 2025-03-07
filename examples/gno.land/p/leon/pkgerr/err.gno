// Package pkgerr provides a custom error wrapper that prepends the realm link to your error
// This is useful for identifying the source package of the error.
//
// Usage:
//
// To wrap an error with realm and chain domain information, use the `New` function:
//
//	err := pkgerr.New(errors.New("my error message")) // in gno.land/r/leon/example
//	fmt.Println(err.Error()) // Output: "r/leon/example: my error message"
package pkgerr

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

// pkgErr is a custom error type that prepends the current
// realm link to the original error message.
type pkgErr struct {
	originalErr error
}

// New creates a new pkgErr with the given error. The returned error will include
// the current realm link in its message.
func New(err error) error {
	if err == nil {
		return nil
	}

	return &pkgErr{originalErr: err}
}

// Error implements the `error` interface for pkgErr.
func (e *pkgErr) Error() string {
	return ufmt.Sprintf("%s: %s",
		strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land/"),
		e.originalErr)
}
