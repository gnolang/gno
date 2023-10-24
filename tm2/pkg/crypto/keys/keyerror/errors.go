package keyerror

import (
	"errors"
	"fmt"
)

const (
	codeKeyNotFound   = 1
	codeWrongPassword = 2
)

type keybaseError interface {
	error
	Code() int
}

type keyNotFoundError struct {
	code int
	name string
}

func (e keyNotFoundError) Code() int {
	return e.code
}

func (e keyNotFoundError) Error() string {
	return fmt.Sprintf("Key %s not found", e.name)
}

// NewErrKeyNotFound returns a standardized error reflecting that the specified key doesn't exist
func NewErrKeyNotFound(name string) error {
	return keyNotFoundError{
		code: codeKeyNotFound,
		name: name,
	}
}

// IsErrKeyNotFound returns true if the given error is keyNotFoundError
func IsErrKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	var keyErr keybaseError
	if errors.As(err, &keyErr) {
		if keyErr.Code() == codeKeyNotFound {
			return true
		}
	}
	return false
}

type wrongPasswordError struct {
	code int
}

func (e wrongPasswordError) Code() int {
	return e.code
}

func (e wrongPasswordError) Error() string {
	return "invalid account password"
}

// NewErrWrongPassword returns a standardized error reflecting that the specified password is wrong
func NewErrWrongPassword() error {
	return wrongPasswordError{
		code: codeWrongPassword,
	}
}

// IsErrWrongPassword returns true if the given error is wrongPasswordError
func IsErrWrongPassword(err error) bool {
	if err == nil {
		return false
	}
	var keyErr keybaseError
	if errors.As(err, &keyErr) {
		if keyErr.Code() == codeWrongPassword {
			return true
		}
	}
	return false
}
