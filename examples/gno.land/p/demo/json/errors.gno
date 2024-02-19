package json

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

type jsonError struct {
	code int
	msg string
}

func (e *jsonError) Error() string {
	return ufmt.Sprintf("Error %d: %s", e.code, e.msg)
}

// error code
const (
	// parser errors (1xx)
	KeyPathNotFound = 100 + iota
	ArrayIndexNotFound
	TokenNotFound
	KeyLevelNotMatched
	NextValueNotFound

	// syntax errors (2xx)
	Overflow = 200 + iota
	EmptyBytes
	InvalidNull
	InvalidBoolean
	InvalidArrayIndex
	InvalidExponents
	NotNumber
	MultipleDecimalPoints

	// type errors (3xx)
	InvalidType = 300 + iota
	InvalidString
	InvalidValue
	InvalidObject
	InvalidArray
	InvalidJSONFormat
	UnknownValueType

	// buffer errors (4xx)
	EofErrorCode = 400 + iota
)

func newJSONError(code int, msg string) error {
	return &jsonError{code, msg}
}