package uint256

import (
	"testing"
)

func shouldEQ(t *testing.T, got, expected any) {
	if got != expected {
		t.Errorf("got %v, expected %v", got, expected)
	}
}

func shouldNEQ(t *testing.T, got, expected any) {
	if got == expected {
		t.Errorf("got %v, didn't expected %v", got, expected)
	}
}

func shouldPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	f()
}

func shouldPanicWithMsg(t *testing.T, f func(), msg string) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		} else {
			if r != msg {
				t.Errorf("excepted panic(%v), got(%v)", msg, r)
			}
		}
	}()
	f()
}

// for original tests
func parseUint(s string) *Uint {
	if len(s) >= 2 && s[:2] == "0x" {
		return MustFromHex(s)
	}
	return MustFromDecimal(s)
}

// for testing.T
func parseUintT(t *testing.T, s string) *Uint {
	return parseUint(s)
}
