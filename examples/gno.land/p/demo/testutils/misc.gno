package testutils

import (
	"testing"
)

// For testing std.GetCallerAt().
func WrapCall(fn func()) {
	fn()
}

func RequirePanic(t *testing.T, fn func(), expectedReason string) {
	defer func() {
		r := recover()
		if r == nil || r.(string) != expectedReason {
			t.Errorf("expected panic with '%v', got '%v'", expectedReason, r)
		}
	}()
	fn()
	t.Errorf("expected panic with '%v', got nothing", expectedReason)
}
