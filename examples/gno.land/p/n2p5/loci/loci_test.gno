package loci

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
)

func TestLociStore(t *testing.T) {
	t.Run("TestSet", func(t *testing.T) {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
		caller := runtime.PreviousRealm()
		store := New()
		// Ensure that the value is nil before setting it.
		if r1 := store.Get(caller.Address()); r1 != nil {
			t.Errorf("expected value to be nil, got '%s'", r1)
		}
		store.Set([]byte("hello"))
		if r2 := store.Get(caller.Address()); string(r2) != "hello" {
			t.Errorf("expected value to be 'hello', got '%s'", r2)
		}
		store.Set([]byte("world"))
		if r3 := store.Get(caller.Address()); string(r3) != "world" {
			t.Errorf("expected value to be 'world', got '%s'", r3)
		}
	})
	t.Run("TestGet", func(t *testing.T) {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
		caller := runtime.PreviousRealm()
		store := New()
		store.Set([]byte("hello"))
		if r0 := store.Get(testutils.TestAddress("nil_user")); r0 != nil {
			t.Errorf("expected value to be nil, got '%s'", r0)
		}
		if r1 := store.Get(caller.Address()); string(r1) != "hello" {
			t.Errorf("expected value to be 'hello', got '%s'", r1)
		}
	})
}
