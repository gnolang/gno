package uassert

import (
	"errors"
	"fmt"
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

func TestPanicsWithError(t *testing.T) {
	mockT := new(mockTestingT)
	if !PanicsWithError(mockT, "panic", func() {
		panic(errors.New("panic"))
	}) {
		t.Error("PanicsWithError should return true")
	}
	mockT.empty(t)

	if PanicsWithError(mockT, "Panic!", func() {
		// noop
	}) {
		t.Error("PanicsWithError should return false")
	}
	mockT.equals(t, "error: func should panic\n\tPanic value:\tnil")

	if PanicsWithError(mockT, "at the disco", func() {
		panic(errors.New("panic"))
	}) {
		t.Error("PanicsWithError should return false")
	}
	mockT.equals(t, "error: func should panic with error:\tpanic\n\tPanic value:\tat the disco")

	if PanicsWithError(mockT, "Panic!", func() {
		panic("panic")
	}) {
		t.Error("PanicsWithError should return false")
	}
	mockT.equals(t, "error: func should panic with error:\tpanic\n\tPanic value:\tPanic!")
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
		//{[]byte("Hello World"), []byte("Hello World"), true, ""},
		// {myType("1"), myType("1"), true, ""},
		//{&struct{}{}, &struct{}{}, true, "pointer equality is based on equality of underlying value"},

		// not expected to be equal
		{"Hello World", 42, false, ""},
		{10, uint(10), false, ""},
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
