package core

import (
	"testing"
)

func TestMatch(t *testing.T) {
	ex := ExpirationNever{}
	switch Expiration(ex).(type) {
	case ExpirationNever:
		t.Log("ExpirationNever")
	default:
		t.Fatalf("expected a match")
	}
}
