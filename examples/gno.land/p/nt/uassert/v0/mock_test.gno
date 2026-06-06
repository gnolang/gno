package uassert_test

import (
	"fmt"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

type mockTestingT struct {
	fmt  string
	args []any
}

// --- interface mock

var _ uassert.TestingT = (*mockTestingT)(nil)

func (mockT *mockTestingT) Helper()                      { /* noop */ }
func (mockT *mockTestingT) Skip(args ...any)             { /* not implmented */ }
func (mockT *mockTestingT) Fail()                        { /* not implmented */ }
func (mockT *mockTestingT) FailNow()                     { /* not implmented */ }
func (mockT *mockTestingT) Logf(fmt string, args ...any) { /* noop */ }

func (mockT *mockTestingT) Fatalf(fmt string, args ...any) {
	mockT.fmt = "fatal: " + fmt
	mockT.args = args
}

func (mockT *mockTestingT) Errorf(fmt string, args ...any) {
	mockT.fmt = "error: " + fmt
	mockT.args = args
}

// --- helpers

func (mockT *mockTestingT) actualString() string {
	res := fmt.Sprintf(mockT.fmt, mockT.args...)
	mockT.reset()
	return res
}

func (mockT *mockTestingT) reset() {
	mockT.fmt = ""
	mockT.args = nil
}

func (mockT *mockTestingT) equals(t *testing.T, expected string) {
	actual := mockT.actualString()

	if expected != actual {
		t.Errorf("mockT differs:\n- expected: %s\n- actual:   %s\n", expected, actual)
	}
}

func (mockT *mockTestingT) empty(t *testing.T) {
	if mockT.fmt != "" || mockT.args != nil {
		actual := mockT.actualString()
		t.Errorf("mockT should be empty, got %s", actual)
	}
}
