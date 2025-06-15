// Package errs provides utilities for combining multiple errors.
// This is a simplified version of https://github.com/uber-go/multierr (MIT License),
// adapted for the Gno programming language with a focus on core functionality
// and idiomatic usage patterns.
//
// Example usage:
//
//	err1 := doSomething()
//	err2 := doSomethingElse()
//	if err := errs.Combine(err1, err2); err != nil {
//	    return err // Returns combined errors or single error
//	}
package errs

import (
	"strings"
)

// multiError represents multiple errors combined into one.
type multiError struct {
	errors []error
}

// Error implements the error interface by returning a single-line representation
// of all contained errors, separated by semicolons.
func (m *multiError) Error() string {
	if m == nil || len(m.errors) == 0 {
		return ""
	}

	var b strings.Builder
	first := true
	for _, err := range m.errors {
		if first {
			first = false
		} else {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

// String returns a multi-line representation of the error.
func (m *multiError) String() string {
	if m == nil || len(m.errors) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("the following errors occurred:")
	for _, err := range m.errors {
		b.WriteString("\n -  ")
		b.WriteString(err.Error())
	}
	return b.String()
}

// Unwrap returns the slice of underlying errors contained in this multiError.
// Returns nil if the receiver is nil.
func (m *multiError) Unwrap() []error {
	if m == nil {
		return nil
	}
	return m.errors
}

// Errors extracts the underlying errors from an error interface.
// If the error is a multiError, it returns its contained errors.
// If the error is nil, returns nil.
// If the error is a regular error, returns a slice containing just that error.
func Errors(err error) []error {
	if err == nil {
		return nil
	}

	if merr, ok := err.(*multiError); ok {
		return merr.Unwrap()
	}

	return []error{err}
}

// Combine merges multiple errors into a single error efficiently.
// It handles several cases:
//   - If all input errors are nil, returns nil
//   - If there's exactly one non-nil error, returns that error directly
//   - If there are multiple non-nil errors, returns a multiError containing them
//   - Flattens any *multiError in the input
func Combine(errs ...error) error {
	nonNil := make([]error, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		if m, ok := err.(*multiError); ok && m != nil {
			nonNil = append(nonNil, m.Unwrap()...)
		} else {
			nonNil = append(nonNil, err)
		}
	}

	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return &multiError{errors: nonNil}
	}
}
