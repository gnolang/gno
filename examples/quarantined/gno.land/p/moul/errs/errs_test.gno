package errs

import (
	"testing"
)

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func newError(msg string) error {
	return &testError{msg: msg}
}

func TestCombine(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected string
	}{
		{
			name:     "single error",
			errors:   []error{newError("error1"), nil},
			expected: "error1",
		},
		{
			name:     "multiple errors",
			errors:   []error{newError("error1"), newError("error2"), newError("error3")},
			expected: "error1; error2; error3",
		},
		{
			name:     "mixed nil and non-nil",
			errors:   []error{nil, newError("error1"), nil, newError("error2")},
			expected: "error1; error2",
		},
		{
			name:     "both nil (2)",
			errors:   []error{nil, nil},
			expected: "",
		},
		{
			name:     "first nil (2)",
			errors:   []error{nil, newError("error2")},
			expected: "error2",
		},
		{
			name:     "second nil (2)",
			errors:   []error{newError("error1"), nil},
			expected: "error1",
		},
		{
			name:     "both non-nil (2)",
			errors:   []error{newError("error1"), newError("error2")},
			expected: "error1; error2",
		},
	}

	for _, tt := range tests {
		err := Combine(tt.errors...)
		if tt.expected == "" {
			if err != nil {
				t.Errorf("%s: expected nil error, got %v", tt.name, err)
			}
			continue
		}

		if err == nil {
			t.Errorf("%s: expected non-nil error", tt.name)
			continue
		}

		if got := err.Error(); got != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, got)
		}
	}
}

func TestErrors(t *testing.T) {
	err1 := newError("error1")
	err2 := newError("error2")
	combined := Combine(err1, err2)

	tests := []struct {
		name          string
		err           error
		expectedCount int
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedCount: 0,
		},
		{
			name:          "single error",
			err:           err1,
			expectedCount: 1,
		},
		{
			name:          "multiple errors",
			err:           combined,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		errs := Errors(tt.err)
		if len(errs) != tt.expectedCount {
			t.Errorf("%s: expected %d errors, got %d", tt.name, tt.expectedCount, len(errs))
		}
	}
}

func TestMultiErrorString(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected string
	}{
		{
			name:     "nil errors",
			errors:   nil,
			expected: "",
		},
		{
			name:     "empty errors",
			errors:   []error{},
			expected: "",
		},
		{
			name:     "single error",
			errors:   []error{newError("error1")},
			expected: "the following errors occurred:\n -  error1",
		},
		{
			name:     "multiple errors",
			errors:   []error{newError("error1"), newError("error2")},
			expected: "the following errors occurred:\n -  error1\n -  error2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merr := &multiError{errors: tt.errors}
			if got := merr.String(); got != tt.expected {
				t.Errorf("multiError.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCombineEdgeCases(t *testing.T) {
	err1 := newError("error1")
	err2 := newError("error2")
	err3 := newError("error3")

	merr := Combine(err1, err2)

	tests := []struct {
		name     string
		errors   []error
		expected string
	}{
		{
			name:     "append to multiError",
			errors:   []error{merr, err3},
			expected: "error1; error2; error3",
		},
		{
			name:     "prepend to multiError",
			errors:   []error{err3, merr},
			expected: "error3; error1; error2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Combine(tt.errors...)
			if got := err.Error(); got != tt.expected {
				t.Errorf("Combine() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestUnwrapNilReceiver(t *testing.T) {
	var merr *multiError
	errs := merr.Unwrap()
	if errs != nil {
		t.Errorf("Expected nil slice for nil receiver, got %v", errs)
	}
}

func TestMultiErrorUnwrap(t *testing.T) {
	var nilMulti *multiError
	if got := nilMulti.Unwrap(); got != nil {
		t.Errorf("Unwrap on nil receiver: expected nil, got %v", got)
	}

	empty := &multiError{errors: nil}
	if got := empty.Unwrap(); got != nil {
		t.Errorf("Unwrap on empty: expected nil, got %v", got)
	}

	single := &multiError{errors: []error{newError("foo")}}
	got := single.Unwrap()
	if len(got) != 1 || got[0].Error() != "foo" {
		t.Errorf("Unwrap on single: expected [foo], got %v", got)
	}

	multi := &multiError{errors: []error{newError("foo"), newError("bar")}}
	got = multi.Unwrap()
	if len(got) != 2 || got[0].Error() != "foo" || got[1].Error() != "bar" {
		t.Errorf("Unwrap on multi: expected [foo bar], got %v", got)
	}
}
