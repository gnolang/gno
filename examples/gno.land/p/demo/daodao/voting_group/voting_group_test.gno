package grc4

import (
	"std"
	"testing"
)

func Test(t *testing.T) {
	{
		admin := "g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a"
		g := grc4.NewGRC4Group(std.Address(admin), nil)
		v := NewGRC4Voting(g)
		got := v.TotalPower()
		expected := uint64(0)
		if got != expected {
			t.Fatalf("expected %q, got %q.", expected, got)
		}
	}
}
