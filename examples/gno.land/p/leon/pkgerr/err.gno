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

	"gno.land/p/nt/ufmt/v0"
)

// pkgErr is a custom error type that prepends the realm link
// (captured at construction time) to the original error message.
type pkgErr struct {
	realmPath   string
	originalErr error
}

// New creates a new pkgErr with the given message. The realm link
// captured by runtime.CurrentRealm() at THIS call's evaluation is
// stored on the returned value, so the prefix in Error() reflects
// the realm that CREATED the error — not whichever realm later
// happens to call .Error() on it.
//
// SECURITY (Class-2-shaped, accepted as display-only): the capture
// uses runtime.CurrentRealm() inside a non-crossing /p/ function and
// is therefore .Title()-vulnerable (the walk resolves to the
// most-recent crossing ancestor, not the immediate caller). For the
// typical package-level-var usage (e.g. `var ErrX = pkgerr.New(...)`
// in r/foo/bar) the construction site IS that package's init, so
// the resolution is correct. For dynamic construction inside a
// function call chain, a contrived chain could produce a misleading
// prefix. This is acceptable here because no consumer in /examples
// uses the prefix for authorization — it is rendered only as part of
// error.Error() text. See docs/resources/gno-security.md for the
// threat-class taxonomy.
func New(msg string) error {
	return &pkgErr{
		realmPath:   runtime.CurrentRealm().PkgPath(),
		originalErr: errors.New(msg),
	}
}

// Wrap behaves like New but reuses an existing error as the cause.
// Same construction-time capture semantics and same SECURITY caveat
// as New.
func Wrap(err error) error {
	if err == nil {
		return nil
	}

	return &pkgErr{
		realmPath:   runtime.CurrentRealm().PkgPath(),
		originalErr: err,
	}
}

func (e pkgErr) Unwrap() error {
	return e.originalErr
}

// Error implements the `error` interface for pkgErr. It uses the
// realmPath captured at construction time so the prefix is stable
// across the error's lifetime, regardless of which realm renders it.
func (e *pkgErr) Error() string {
	return ufmt.Sprintf("%s: %s",
		strings.TrimPrefix(e.realmPath, "gno.land/"),
		e.originalErr)
}
