package uassert

import (
	"errors"
	"fmt"
	"std"
	"testing"
)

var _ TestingT = (*testing.T)(nil)

func TestMock(t *testing.T) {
	mockT := new(mockTestingT)
	mockT.empty(t)
	NoError(mockT, errors.New("foo"))
	mockT.equals(t, "error: unexpected error: foo")
	NoError(mockT, errors.New("foo"), "custom message")
	mockT.equals(t, "error: unexpected error: foo - custom message")
	NoError(mockT, errors.New("foo"), "custom", "message")
	mockT.equals(t, "error: unexpected error: foo - custom message")
}

func TestNoError(t *testing.T) {
	mockT := new(mockTestingT)
	True(t, NoError(mockT, nil))
	mockT.empty(t)
	False(t, NoError(mockT, errors.New("foo bar")))
	mockT.equals(t, "error: unexpected error: foo bar")
}

func TestError(t *testing.T) {
	mockT := new(mockTestingT)
	True(t, Error(mockT, errors.New("foo bar")))
	mockT.empty(t)
	False(t, Error(mockT, nil))
	mockT.equals(t, "error: an error is expected but got nil")
}

func TestErrorContains(t *testing.T) {
	mockT := new(mockTestingT)

	// nil error
	var err error
	False(t, ErrorContains(mockT, err, ""), "ErrorContains should return false for nil arg")
}

func TestTrue(t *testing.T) {
	mockT := new(mockTestingT)
	if !True(mockT, true) {
		t.Error("True should return true")
	}
	mockT.empty(t)
	if True(mockT, false) {
		t.Error("True should return false")
	}
	mockT.equals(t, "error: should be true")
}

func TestFalse(t *testing.T) {
	mockT := new(mockTestingT)
	if !False(mockT, false) {
		t.Error("False should return true")
	}
	mockT.empty(t)
	if False(mockT, true) {
		t.Error("False should return false")
	}
	mockT.equals(t, "error: should be false")
}

func TestPanicsWithMessage(t *testing.T) {
	mockT := new(mockTestingT)
	if !PanicsWithMessage(mockT, "panic", func() {
		panic(errors.New("panic"))
	}) {
		t.Error("PanicsWithMessage should return true")
	}
	mockT.empty(t)

	if PanicsWithMessage(mockT, "Panic!", func() {
		// noop
	}) {
		t.Error("PanicsWithMessage should return false")
	}
	mockT.equals(t, "error: func should panic\n\tPanic value:\tnil")

	if PanicsWithMessage(mockT, "at the disco", func() {
		panic(errors.New("panic"))
	}) {
		t.Error("PanicsWithMessage should return false")
	}
	mockT.equals(t, "error: func should panic with message:\tat the disco\n\tPanic value:\tpanic")

	if PanicsWithMessage(mockT, "Panic!", func() {
		panic("panic")
	}) {
		t.Error("PanicsWithMessage should return false")
	}
	mockT.equals(t, "error: func should panic with message:\tPanic!\n\tPanic value:\tpanic")
}

func TestNotPanics(t *testing.T) {
	mockT := new(mockTestingT)

	if !NotPanics(mockT, func() {
		// noop
	}) {
		t.Error("NotPanics should return true")
	}
	mockT.empty(t)

	if NotPanics(mockT, func() {
		panic("Panic!")
	}) {
		t.Error("NotPanics should return false")
	}
}

func TestEqual(t *testing.T) {
	mockT := new(mockTestingT)

	cases := []struct {
		expected interface{}
		actual   interface{}
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
		{std.Address("g12345"), std.Address("g12345"), true, ""},
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
			res := Equal(mockT, c.expected, c.actual)

			if res != c.result {
				t.Errorf("%s should return %v: %s - %s", name, c.result, c.remark, mockT.actualString())
			}
		})
	}
}

func TestNotEqual(t *testing.T) {
	mockT := new(mockTestingT)

	cases := []struct {
		expected interface{}
		actual   interface{}
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
		{std.Address("g12345"), std.Address("g67890"), true, ""},
		// XXX: continue

		// not expected to be not equal
		{"Hello World", "Hello World", false, ""},
		{123, 123, false, ""},
		{123.5, 123.5, false, ""},
		{nil, nil, false, ""},
		{int32(123), int32(123), false, ""},
		{uint64(123), uint64(123), false, ""},
		{std.Address("g12345"), std.Address("g12345"), false, ""},
		// XXX: continue

		// expected to raise errors
		// XXX: todo
	}

	for _, c := range cases {
		name := fmt.Sprintf("NotEqual(%v, %v)", c.expected, c.actual)
		t.Run(name, func(t *testing.T) {
			res := NotEqual(mockT, c.expected, c.actual)

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
		obj           interface{}
		expectedEmpty bool
	}{
		// expected to be empty
		{"", true},
		{0, true},
		{int(0), true},
		{int64(0), true},
		{uint(0), true},
		// XXX: continue

		// not expected to be empty
		{"Hello World", false},
		{1, false},
		{int32(1), false},
		{uint64(1), false},
		{std.Address("g12345"), false},

		// unsupported
		{nil, false},
		{myStruct{}, false},
		{&myStruct{}, false},
	}

	for _, c := range cases {
		name := fmt.Sprintf("Empty(%v)", c.obj)
		t.Run(name, func(t *testing.T) {
			res := Empty(mockT, c.obj)

			if res != c.expectedEmpty {
				t.Errorf("%s should return %v: %s", name, c.expectedEmpty, mockT.actualString())
			}
		})
	}
}

func TestNil(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Nil(mockT, nil))
	mockT.empty(t)

	False(t, Nil(mockT, "not nil"))
	mockT.equals(t, "error: uassert.Nil: object should be nil")
}

func TestNotNil(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, NotNil(mockT, "not nil"))
	mockT.empty(t)

	False(t, NotNil(mockT, nil))
	mockT.equals(t, "error: uassert.NotNil: object should not be nil")
}

func TestContains(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Contains(mockT, "Hello World", "World"))
	mockT.empty(t)

	False(t, Contains(mockT, "Hello World", "Foo"))
	mockT.equals(t, "error: uassert.Contains: \"Hello World\" does not contain \"Foo\"")

	True(t, Contains(mockT, []string{"foo", "bar"}, "foo"))
	mockT.empty(t)

	False(t, Contains(mockT, []string{"foo", "bar"}, "baz"))
	mockT.equals(t, "error: uassert.Contains: [foo bar] does not contain baz")
}

func TestLen(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Len(mockT, "Hello", 5))
	mockT.empty(t)

	False(t, Len(mockT, "Hello", 6))
	mockT.equals(t, "error: uassert.Len: expected length 6 but got 5")

	True(t, Len(mockT, []interface{}{"foo", "bar"}, 2))
	mockT.empty(t)

	False(t, Len(mockT, []interface{}{"foo", "bar"}, 3))
	mockT.equals(t, "error: uassert.Len: expected length 3 but got 2")
}

func TestGreater(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Greater(mockT, 2, 1))
	mockT.empty(t)

	False(t, Greater(mockT, 1, 2))
	mockT.equals(t, "error: uassert.Greater: 1 is not greater than 2")
}

func TestLess(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Less(mockT, 1, 2))
	mockT.empty(t)

	False(t, Less(mockT, 2, 1))
	mockT.equals(t, "error: uassert.Less: 2 is not less than 1")
}

func TestIsType(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, IsType(mockT, "string", "test"))
	mockT.empty(t)

	False(t, IsType(mockT, "string", 123))
	mockT.equals(t, "error: uassert.IsType: expected type string but got int")
}

func TestRegexp(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Regexp(mockT, "^[a-z]+$", "test"))
	mockT.empty(t)

	False(t, Regexp(mockT, "^[0-9]+$", "test"))
	mockT.equals(t, "error: uassert.Regexp: \"test\" does not match \"^[0-9]+$\"")
}

func TestGreaterOrEqual(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, GreaterOrEqual(mockT, 2, 1))
	mockT.empty(t)

	True(t, GreaterOrEqual(mockT, 2, 2))
	mockT.empty(t)

	False(t, GreaterOrEqual(mockT, 1, 2))
	mockT.equals(t, "error: uassert.GreaterOrEqual: 1 is not greater than or equal to 2")
}

func TestLessOrEqual(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, LessOrEqual(mockT, 1, 2))
	mockT.empty(t)

	True(t, LessOrEqual(mockT, 2, 2))
	mockT.empty(t)

	False(t, LessOrEqual(mockT, 2, 1))
	mockT.equals(t, "error: uassert.LessOrEqual: 2 is not less than or equal to 1")
}

func TestNotContains(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, NotContains(mockT, "Hello World", "Foo"))
	mockT.empty(t)

	False(t, NotContains(mockT, "Hello World", "World"))
	mockT.equals(t, "error: uassert.NotContains: \"Hello World\" contains \"World\"")

	True(t, NotContains(mockT, []string{"foo", "bar"}, "baz"))
	mockT.empty(t)

	False(t, NotContains(mockT, []string{"foo", "bar"}, "foo"))
	mockT.equals(t, "error: uassert.NotContains: [foo bar] contains foo")
}

func TestElementsMatch(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, ElementsMatch(mockT, []interface{}{"foo", "bar"}, []interface{}{"bar", "foo"}))
	mockT.empty(t)

	False(t, ElementsMatch(mockT, []interface{}{"foo", "bar"}, []interface{}{"baz", "foo"}))
	mockT.equals(t, "error: uassert.ElementsMatch: lists do not match")
}

func TestSubset(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Subset(mockT, []interface{}{"foo", "bar", "baz"}, []interface{}{"foo", "bar"}))
	mockT.empty(t)

	False(t, Subset(mockT, []interface{}{"foo", "bar"}, []interface{}{"foo", "baz"}))
	mockT.equals(t, "error: uassert.Subset: item baz not found in set")
}

func TestNotSubset(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, NotSubset(mockT, []interface{}{"foo", "bar"}, []interface{}{"foo", "baz"}))
	mockT.empty(t)

	False(t, NotSubset(mockT, []interface{}{"foo", "bar"}, []interface{}{"foo"}))
	mockT.equals(t, "error: uassert.NotSubset: subset is fully present in set")
}

func TestIsInstanceOf(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, IsInstanceOf(mockT, "string", "test"))
	mockT.empty(t)

	False(t, IsInstanceOf(mockT, "string", 123))
	mockT.equals(t, "error: uassert.IsInstanceOf: 123 is not an instance of string")
}

func TestSame(t *testing.T) {
	mockT := new(mockTestingT)

	obj := &struct{}{}
	True(t, Same(mockT, obj, obj))
	mockT.empty(t)

	False(t, Same(mockT, obj, &struct{}{}))
	mockT.equals(t, "error: uassert.Same: expected and actual do not reference the same object")
}

func TestNotSame(t *testing.T) {
	mockT := new(mockTestingT)

	obj := &struct{}{}
	True(t, NotSame(mockT, obj, &struct{}{}))
	mockT.empty(t)

	False(t, NotSame(mockT, obj, obj))
	mockT.equals(t, "error: uassert.NotSame: expected and actual reference the same object")
}

func TestPanics(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, Panics(mockT, func() { panic("test panic") }))
	mockT.empty(t)

	False(t, Panics(mockT, func() {}))
	mockT.equals(t, "error: uassert.Panics: function did not panic")
}

func TestPanicsWithValue(t *testing.T) {
	mockT := new(mockTestingT)

	True(t, PanicsWithValue(mockT, "test panic", func() { panic("test panic") }))
	mockT.empty(t)

	False(t, PanicsWithValue(mockT, "expected panic", func() { panic("test panic") }))
	mockT.equals(t, "error: uassert.PanicsWithValue: panic value test panic does not equal expected expected panic")
}

func TestNotEmpty(t *testing.T) {
	mockT := new(mockTestingT)

	cases := []struct {
		obj              interface{}
		expectedNotEmpty bool
	}{
		// expected to be empty
		{"", false},
		{0, false},
		{int(0), false},
		{int64(0), false},
		{uint(0), false},
		{std.Address(""), false},

		// not expected to be empty
		{"Hello World", true},
		{1, true},
		{int32(1), true},
		{uint64(1), true},
		{std.Address("g12345"), true},

		// unsupported
		{nil, false},
		{myStruct{}, false},
		{&myStruct{}, false},
	}

	for _, c := range cases {
		name := fmt.Sprintf("NotEmpty(%v)", c.obj)
		t.Run(name, func(t *testing.T) {
			res := NotEmpty(mockT, c.obj)

			if res != c.expectedNotEmpty {
				t.Errorf("%s should return %v: %s", name, c.expectedNotEmpty, mockT.actualString())
			}
		})
	}
}
