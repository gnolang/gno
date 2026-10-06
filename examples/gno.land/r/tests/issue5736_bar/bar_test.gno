package issue5736_bar

import (
	"testing"
)

func TestCall(cur realm, t *testing.T) {
	got := Call(cross(cur))
	if got != "246" {
		t.Fatalf("got %s, want 246", got)
	}
}
