package conversion

import (
	"testing"
)

func shouldEqual(t *testing.T, got, want interface{}) {
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}


func shouldPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}