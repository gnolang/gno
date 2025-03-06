package pkgerr

import (
	"errors"
	"testing"
)

func TestNew_ErrorMessage(t *testing.T) {
	origErr := errors.New("my error message")
	err := New(origErr)
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Error() == "" {
		t.Errorf("expected non-empty error message, got empty string")
	}
}

func TestNew_NilError(t *testing.T) {
	err := New(nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestUnwrap(t *testing.T) {
	origErr := errors.New("original error")
	err := New(origErr)

	err.Error()
}
