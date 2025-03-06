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

// PkgErr is a custom error type that prepends the current
// realm link to the original error message.
type PkgErr struct {
	originalErr error
}

// New creates a new PkgErr with the given error. The returned error will include
// the current realm link in its message.
func New(err error) error {
	if err == nil {
		return nil
	}

	return &PkgErr{originalErr: err}
}

// Error implements the `error` interface for PkgErr.
func (e *PkgErr) Error() string {
	return ufmt.Sprintf("%s: %s",
		strings.TrimPrefix(std.CurrentRealm().PkgPath(), "gno.land/"),
		e.originalErr)
}
