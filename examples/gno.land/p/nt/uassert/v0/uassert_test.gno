package uassert_test

import (
	"errors"
	"fmt"
	"testing"

	"gno.land/p/nt/uassert/v0"
	tests "gno.land/r/tests/vm"
)

var _ uassert.TestingT = (*testing.T)(nil)

func TestMock(t *testing.T) {
	mockT := new(mockTestingT)
	mockT.empty(t)
	uassert.NoError(mockT, errors.New("foo"))
	mockT.equals(t, "error: unexpected error: foo")
	uassert.NoError(mockT, errors.New("foo"), "custom message")
	mockT.equals(t, "error: unexpected error: foo - custom message")
	uassert.NoError(mockT, errors.New("foo"), "custom", "message")
	mockT.equals(t, "error: unexpected error: foo - custom message")
}

func TestNoError(t *testing.T) {
	mockT := new(mockTestingT)
	uassert.True(t, uassert.NoError(mockT, nil))
	mockT.empty(t)
	uassert.False(t, uassert.NoError(mockT, errors.New("foo bar")))
	mockT.equals(t, "error: unexpected error: foo bar")
}

func TestError(t *testing.T) {
	mockT := new(mockTestingT)
	uassert.True(t, uassert.Error(mockT, errors.New("foo bar")))
	mockT.empty(t)
	uassert.False(t, uassert.Error(mockT, nil))
	mockT.equals(t, "error: an error is expected but got nil")
}

func TestErrorContains(t *testing.T) {
	mockT := new(mockTestingT)

	// nil error
	var err error
	uassert.False(t, uassert.ErrorContains(mockT, err, ""), "ErrorContains should return false for nil arg")
}

func TestTrue(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.True(mockT, true) {
		t.Error("True should return true")
	}
	mockT.empty(t)
	if uassert.True(mockT, false) {
		t.Error("True should return false")
	}
	mockT.equals(t, "error: should be true")
}

func TestFalse(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.False(mockT, false) {
		t.Error("False should return true")
	}
	mockT.empty(t)
	if uassert.False(mockT, true) {
		t.Error("False should return false")
	}
	mockT.equals(t, "error: should be false")
}

func TestPanicsWithMessage(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.PanicsWithMessage(mockT, "panic", func() {
		panic(errors.New("panic"))
	}) {
		t.Error("PanicsWithMessage should return true")
	}
	mockT.empty(t)

	if uassert.PanicsWithMessage(mockT, "Panic!", func() {
		// noop
	}) {
		t.Error("PanicsWithMessage should return false")
	}
	mockT.equals(t, "error: func should panic\n\tPanic value:\tnil")

	if uassert.PanicsWithMessage(mockT, "at the disco", func() {
		panic(errors.New("panic"))
	}) {
		t.Error("PanicsWithMessage should return false")
	}
	mockT.equals(t, "error: func should panic with message:\t\"at the disco\"\n\tActual panic value:\t\"panic\"")

	if uassert.PanicsWithMessage(mockT, "Panic!", func() {
		panic("panic")
	}) {
		t.Error("PanicsWithMessage should return false")
	}
	mockT.equals(t, "error: func should panic with message:\t\"Panic!\"\n\tActual panic value:\t\"panic\"")
}

func TestPanicsContains(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.PanicsContains(mockT, "panic", func() {
		panic(errors.New("panic: something happened"))
	}) {
		t.Error("PanicsContains should return true for substring match")
	}
	mockT.empty(t)

	if uassert.PanicsContains(mockT, "notfound", func() {
		panic(errors.New("panic: something happened"))
	}) {
		t.Error("PanicsContains should return false for missing substring")
	}
	mockT.equals(t, "error: func should panic with message containing:\t\"notfound\"\n\tActual panic value:\t\"panic: something happened\"")

	if uassert.PanicsContains(mockT, "panic", func() {
		// noop
	}) {
		t.Error("PanicsContains should return false when no panic occurs")
	}
	mockT.equals(t, "error: func should panic\n\tPanic value:\tnil")
}

func TestAbortsWithMessage(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.AbortsWithMessage(mockT, "abort message", func() {
		tests.ExecSwitch(cross, func() {
			panic("abort message")
		})
		panic("dontcare")
	}) {
		t.Error("AbortsWithMessage should return true")
	}
	mockT.empty(t)

	if uassert.AbortsWithMessage(mockT, "Abort!", func() {
		// noop
	}) {
		t.Error("AbortsWithMessage should return false")
	}
	mockT.equals(t, "error: func should abort")

	if uassert.AbortsWithMessage(mockT, "at the disco", func() {
		tests.ExecSwitch(cross, func() {
			panic("abort message")
		})
		panic("dontcare")
	}) {
		t.Error("AbortsWithMessage should return false (wrong message)")
	}
	mockT.equals(t, "error: func should abort with message:\t\"at the disco\"\n\tActual abort value:\t\"abort message\"")

	// Test that non-crossing panics don't count as abort.
	uassert.PanicsWithMessage(mockT, "non-abort panic", func() {
		uassert.AbortsWithMessage(mockT, "dontcare2", func() {
			panic("non-abort panic")
		})
		t.Error("AbortsWithMessage should not have caught non-abort panic")
	}, "non-abort panic")
	mockT.empty(t)

	// Test case where abort value is not a string
	if uassert.AbortsWithMessage(mockT, "doesn't matter", func() {
		tests.ExecSwitch(cross, func() {
			panic(123) // abort with an integer
		})
		panic("dontcare")
	}) {
		t.Error("AbortsWithMessage should return false when abort value is not a string")
	}
	mockT.equals(t, "error: func should abort with message:\t\"doesn't matter\"\n\tActual abort value:\t\"123\"")

	// XXX: test with Error
}

func TestAbortsContains(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.AbortsContains(mockT, "abort", func() {
		tests.ExecSwitch(cross, func() {
			panic("abort message: something happened")
		})
		panic("dontcare")
	}) {
		t.Error("AbortsContains should return true for substring match")
	}
	mockT.empty(t)

	if uassert.AbortsContains(mockT, "notfound", func() {
		tests.ExecSwitch(cross, func() {
			panic("abort message: something happened")
		})
		panic("dontcare")
	}) {
		t.Error("AbortsContains should return false for missing substring")
	}
	mockT.equals(t, "error: func should abort with message containing:\t\"notfound\"\n\tActual abort value:\t\"abort message: something happened\"")

	if uassert.AbortsContains(mockT, "abort", func() {
		// noop
	}) {
		t.Error("AbortsContains should return false when no abort occurs")
	}
	mockT.equals(t, "error: func should abort")
}

func TestNotAborts(t *testing.T) {
	mockT := new(mockTestingT)

	if !uassert.NotPanics(mockT, func() {
		// noop
	}) {
		t.Error("NotAborts should return true")
	}
	mockT.empty(t)

	if uassert.NotPanics(mockT, func() {
		tests.ExecSwitch(cross, func() {
			panic("Abort!")
		})
		panic("dontcare")
	}) {
		t.Error("NotAborts should return false")
	}
	mockT.equals(t, "error: func should not abort\n\tAbort value:\tAbort!")
}

func TestNotPanics(t *testing.T) {
	mockT := new(mockTestingT)

	if !uassert.NotPanics(mockT, func() {
		// noop
	}) {
		t.Error("NotPanics should return true")
	}
	mockT.empty(t)

	if uassert.NotPanics(mockT, func() {
		panic("Panic!")
	}) {
		t.Error("NotPanics should return false")
	}
	mockT.equals(t, "error: func should not panic\n\tPanic value:\tPanic!")
}

func TestEqual(t *testing.T) {
	mockT := new(mockTestingT)

	cases := []struct {
		expected any
		actual   any
		result   bool
		remark   string
	}{
		// expected to be equal
		{"Hello World", "Hello World", true, ""},
		{123, 123, true, ""},
		{123.5, 123.5, true, ""},
		{nil, nil, true, ""},
		{int32(123), int32(123), true, ""},
		{uint64(123), uint64(123), true, ""},
		{address("g12345"), address("g12345"), true, ""},
		// XXX: continue

		// not expected to be equal
		{"Hello World", 42, false, ""},
		{41, 42, false, ""},
		{10, uint(10), false, ""},
		// XXX: continue

		// expected to raise errors
		// XXX: todo
	}

	for _, c := range cases {
		name := fmt.Sprintf("Equal(%v, %v)", c.expected, c.actual)
		t.Run(name, func(t *testing.T) {
			res := uassert.Equal(mockT, c.expected, c.actual)

			if res != c.result {
				t.Errorf("%s should return %v: %s - %s", name, c.result, c.remark, mockT.actualString())
			}
		})
	}
}

func TestNotEqual(t *testing.T) {
	mockT := new(mockTestingT)

	cases := []struct {
		expected any
		actual   any
		result   bool
		remark   string
	}{
		// expected to be not equal
		{"Hello World", "Hello", true, ""},
		{123, 124, true, ""},
		{123.5, 123.6, true, ""},
		{nil, 123, true, ""},
		{int32(123), int32(124), true, ""},
		{uint64(123), uint64(124), true, ""},
		{address("g12345"), address("g67890"), true, ""},
		// XXX: continue

		// not expected to be not equal
		{"Hello World", "Hello World", false, ""},
		{123, 123, false, ""},
		{123.5, 123.5, false, ""},
		{nil, nil, false, ""},
		{int32(123), int32(123), false, ""},
		{uint64(123), uint64(123), false, ""},
		{address("g12345"), address("g12345"), false, ""},
		// XXX: continue

		// expected to raise errors
		// XXX: todo
	}

	for _, c := range cases {
		name := fmt.Sprintf("NotEqual(%v, %v)", c.expected, c.actual)
		t.Run(name, func(t *testing.T) {
			res := uassert.NotEqual(mockT, c.expected, c.actual)

			if res != c.result {
				t.Errorf("%s should return %v: %s - %s", name, c.result, c.remark, mockT.actualString())
			}
		})
	}
}

type myStruct struct {
	S string
	I int
}

func TestEmpty(t *testing.T) {
	mockT := new(mockTestingT)

	cases := []struct {
		obj           any
		expectedEmpty bool
	}{
		// expected to be empty
		{"", true},
		{0, true},
		{int(0), true},
		{int32(0), true},
		{int64(0), true},
		{uint(0), true},
		// XXX: continue

		// not expected to be empty
		{"Hello World", false},
		{1, false},
		{int32(1), false},
		{uint64(1), false},
		{address("g12345"), false},

		// unsupported
		{nil, false},
		{myStruct{}, false},
		{&myStruct{}, false},
	}

	for _, c := range cases {
		name := fmt.Sprintf("Empty(%v)", c.obj)
		t.Run(name, func(t *testing.T) {
			res := uassert.Empty(mockT, c.obj)

			if res != c.expectedEmpty {
				t.Errorf("%s should return %v: %s", name, c.expectedEmpty, mockT.actualString())
			}
		})
	}
}

func TestEqualWithStringDiff(t *testing.T) {
	cases := []struct {
		name        string
		expected    string
		actual      string
		shouldPass  bool
		expectedMsg string
	}{
		{
			name:        "Identical strings",
			expected:    "Hello, world!",
			actual:      "Hello, world!",
			shouldPass:  true,
			expectedMsg: "",
		},
		{
			name:        "Different strings - simple",
			expected:    "Hello, world!",
			actual:      "Hello, World!",
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: strings are different\n\tDiff: Hello, [-w][+W]orld!",
		},
		{
			name:        "Different strings - complex",
			expected:    "The quick brown fox jumps over the lazy dog",
			actual:      "The quick brown cat jumps over the lazy dog",
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: strings are different\n\tDiff: The quick brown [-fox][+cat] jumps over the lazy dog",
		},
		{
			name:        "Different strings - prefix",
			expected:    "prefix_string",
			actual:      "string",
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: strings are different\n\tDiff: [-prefix_]string",
		},
		{
			name:        "Different strings - suffix",
			expected:    "string",
			actual:      "string_suffix",
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: strings are different\n\tDiff: string[+_suffix]",
		},
		{
			name:        "Empty string vs non-empty string",
			expected:    "",
			actual:      "non-empty",
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: strings are different\n\tDiff: [+non-empty]",
		},
		{
			name:        "Non-empty string vs empty string",
			expected:    "non-empty",
			actual:      "",
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: strings are different\n\tDiff: [-non-empty]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			result := uassert.Equal(mockT, tc.expected, tc.actual)

			if result != tc.shouldPass {
				t.Errorf("Expected Equal to return %v, but got %v", tc.shouldPass, result)
			}

			if tc.shouldPass {
				mockT.empty(t)
			} else {
				mockT.equals(t, tc.expectedMsg)
			}
		})
	}
}

func TestNotEmpty(t *testing.T) {
	mockT := new(mockTestingT)

	cases := []struct {
		obj              any
		expectedNotEmpty bool
	}{
		// expected to be empty
		{"", false},
		{0, false},
		{int(0), false},
		{int32(0), false},
		{int64(0), false},
		{uint(0), false},
		{address(""), false},

		// not expected to be empty
		{"Hello World", true},
		{1, true},
		{int32(1), true},
		{uint64(1), true},
		{address("g12345"), true},

		// unsupported
		{nil, false},
		{myStruct{}, false},
		{&myStruct{}, false},
	}

	for _, c := range cases {
		name := fmt.Sprintf("NotEmpty(%v)", c.obj)
		t.Run(name, func(t *testing.T) {
			res := uassert.NotEmpty(mockT, c.obj)

			if res != c.expectedNotEmpty {
				t.Errorf("%s should return %v: %s", name, c.expectedNotEmpty, mockT.actualString())
			}
		})
	}
}

func TestNil(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.Nil(mockT, nil) {
		t.Error("Nil should return true")
	}
	mockT.empty(t)
	if uassert.Nil(mockT, 0) {
		t.Error("Nil should return false")
	}
	mockT.equals(t, "error: should be nil")
	if uassert.Nil(mockT, (*int)(nil)) {
		t.Error("Nil should return false")
	}
	mockT.equals(t, "error: should be nil")
}

func TestNotNil(t *testing.T) {
	mockT := new(mockTestingT)
	if uassert.NotNil(mockT, nil) {
		t.Error("NotNil should return false")
	}
	mockT.equals(t, "error: should not be nil")
	if !uassert.NotNil(mockT, 0) {
		t.Error("NotNil should return true")
	}
	mockT.empty(t)
	if !uassert.NotNil(mockT, (*int)(nil)) {
		t.Error("NotNil should return true")
	}
	mockT.empty(t)
}

func TestTypedNil(t *testing.T) {
	mockT := new(mockTestingT)
	if uassert.TypedNil(mockT, nil) {
		t.Error("TypedNil should return false")
	}
	mockT.equals(t, "error: should be typed-nil but got nil instead")
	if uassert.TypedNil(mockT, 0) {
		t.Error("TypedNil should return false")
	}
	mockT.equals(t, "error: should be typed-nil")
	if !uassert.TypedNil(mockT, (*int)(nil)) {
		t.Error("TypedNil should return true")
	}
	mockT.empty(t)
}

func TestNotTypedNil(t *testing.T) {
	mockT := new(mockTestingT)
	if !uassert.NotTypedNil(mockT, nil) {
		t.Error("NotTypedNil should return true")
	}
	mockT.empty(t)
	if !uassert.NotTypedNil(mockT, 0) {
		t.Error("NotTypedNil should return true")
	}
	mockT.empty(t)
	if uassert.NotTypedNil(mockT, (*int)(nil)) {
		t.Error("NotTypedNil should return false")
	}
	mockT.equals(t, "error: should not be typed-nil")
}
