package uassert

import (
	"fmt"
	"testing"
)

type mockTestingT struct {
	fmt  string
	args []interface{}
}

// --- interface mock

var _ TestingT = (*mockTestingT)(nil)

func (mockT *mockTestingT) Helper()                              { /* noop */ }
func (mockT *mockTestingT) Skip(args ...interface{})             { /* not implmented */ }
func (mockT *mockTestingT) Fail()                                { /* not implmented */ }
func (mockT *mockTestingT) FailNow()                             { /* not implmented */ }
func (mockT *mockTestingT) Logf(fmt string, args ...interface{}) { /* noop */ }

func (mockT *mockTestingT) Fatalf(fmt string, args ...interface{}) {
	mockT.fmt = "fatal: " + fmt
	mockT.args = args
}

func (mockT *mockTestingT) Errorf(fmt string, args ...interface{}) {
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
