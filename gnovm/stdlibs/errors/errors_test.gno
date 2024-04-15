// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors_test

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewEqual(t *testing.T) {
	// Different allocations should not be equal.
	if errors.New("abc") == errors.New("abc") {
		t.Errorf(`New("abc") == New("abc")`)
	}
	if errors.New("abc") == errors.New("xyz") {
		t.Errorf(`New("abc") == New("xyz")`)
	}

	// Same allocation should be equal to itself (not crash).
	err := errors.New("jkl")
	if err != err {
		t.Errorf(`err != err`)
	}
}

func TestErrorMethod(t *testing.T) {
	err := errors.New("abc")
	if err.Error() != "abc" {
		t.Errorf(`New("abc").Error() = %q, want %q`, err.Error(), "abc")
	}
}

func ExampleNew() {
	err := errors.New("emit macho dwarf: elf header corrupted")
	if err != nil {
		fmt.Print(err)
	}
	// Output: emit macho dwarf: elf header corrupted
}

// The fmt package's Errorf function lets us use the package's formatting
// features to create descriptive error messages.
func ExampleNew_errorf() {
	const name, id = "bimmler", 17
	err := fmt.Errorf("user %q (id %d) not found", name, id)
	if err != nil {
		fmt.Print(err)
	}
	// Output: user "bimmler" (id 17) not found
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func (e *customError) Is(target error) bool {
	t, ok := target.(*customError)
	if !ok {
		return false
	}
	return e.msg == t.msg
}

func TestIs(t *testing.T) {
	testCases := []struct {
		err    error
		target error
		want   bool
	}{
		{nil, nil, true},
		{errors.New("test"), nil, false},
		{&customError{msg: "custom"}, &customError{msg: "custom"}, true},
		{&customError{msg: "custom"}, &customError{msg: "other"}, false},
		{&customError{msg: "custom"}, errors.New("custom"), false},
		{errors.New("custom"), &customError{msg: "custom"}, false},
		{errors.New("custom"), errors.New("custom"), true},
		{&wrappedError{err: errors.New("wrapped")}, errors.New("wrapped"), true},
		{&wrappedError{err: &customError{msg: "custom"}}, &customError{msg: "custom"}, true},
		{&wrappedError{err: errors.New("wrapped")}, errors.New("other"), false},
		{&wrappedError{err: &customError{msg: "custom"}}, &customError{msg: "other"}, false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test Case %d", i+1), func(t *testing.T) {
			got := errors.Is(tc.err, tc.target)
			if got != tc.want {
				t.Errorf("Is(%v, %v) = %v, want %v", tc.err, tc.target, got, tc.want)
			}
		})
	}
}

type wrappedError struct {
	err error
}

func (e *wrappedError) Error() string {
	return "wrapped error: " + e.err.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.err
}
