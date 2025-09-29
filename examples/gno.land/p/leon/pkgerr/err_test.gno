package pkgerr

import (
	"errors"
	"strings"
	"testing"
)

const pkgPath = "gno.land/r/leon/test"

var prefix = strings.TrimPrefix(pkgPath, "gno.land/")

func TestNew(t *testing.T) {
	err := New("my error message")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	testing.SetRealm(testing.NewCodeRealm(pkgPath))

	expected := prefix + ": my error message"

	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestWrap(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(pkgPath))

	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr)
	if wrappedErr == nil {
		t.Fatal("Expected an error, got nil")
	}

	expected := prefix + ": original error"
	if wrappedErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, wrappedErr.Error())
	}
}

func TestUnwrap(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(pkgPath))
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr)

	unwrappedErr := wrappedErr.(*pkgErr).Unwrap()
	if unwrappedErr != originalErr {
		t.Errorf("Expected unwrapped error %v, got %v", originalErr, unwrappedErr)
	}
}

func TestErrorMethod(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(pkgPath))
	originalErr := errors.New("original error")
	pkgErr := &pkgErr{originalErr: originalErr}

	expected := prefix + ": original error"
	if pkgErr.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, pkgErr.Error())
	}
}

func TestWrapNilError(t *testing.T) {
	err := Wrap(nil)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestNewWithEmptyMessage(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(pkgPath))
	err := New("")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	expected := prefix + ": "
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}
