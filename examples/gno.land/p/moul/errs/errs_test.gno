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
			name:     "nil errors",
			errors:   []error{nil, nil},
			expected: "",
		},
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

func TestAppend(t *testing.T) {
	tests := []struct {
		name     string
		err1     error
		err2     error
		expected string
	}{
		{
			name:     "both nil",
			err1:     nil,
			err2:     nil,
			expected: "",
		},
		{
			name:     "first nil",
			err1:     nil,
			err2:     newError("error2"),
			expected: "error2",
		},
		{
			name:     "second nil",
			err1:     newError("error1"),
			err2:     nil,
			expected: "error1",
		},
		{
			name:     "both non-nil",
			err1:     newError("error1"),
			err2:     newError("error2"),
			expected: "error1; error2",
		},
	}

	for _, tt := range tests {
		err := Append(tt.err1, tt.err2)
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

func TestAppendEdgeCases(t *testing.T) {
	err1 := newError("error1")
	err2 := newError("error2")
	err3 := newError("error3")

	// Create a multiError
	merr := Combine(err1, err2)

	tests := []struct {
		name     string
		err1     error
		err2     error
		expected string
	}{
		{
			name:     "append to multiError",
			err1:     merr,
			err2:     err3,
			expected: "error1; error2; error3",
		},
		{
			name:     "prepend to multiError",
			err1:     err3,
			err2:     merr,
			expected: "error3; error1; error2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Append(tt.err1, tt.err2)
			if got := err.Error(); got != tt.expected {
				t.Errorf("Append() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestErrorsNilReceiver(t *testing.T) {
	var merr *multiError
	errs := merr.Errors()
	if errs != nil {
		t.Errorf("Expected nil slice for nil receiver, got %v", errs)
	}
}
