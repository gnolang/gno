package loci

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
)

func TestLociStore(cur realm, t *testing.T) {
	caller := testutils.TestAddress("caller")

	t.Run("TestSet", func(cur realm, t *testing.T) {
		testing.SetRealm(testing.NewUserRealm(caller))
		store := New()
		// Ensure that the value is nil before setting it.
		if r1 := store.Get(caller); r1 != nil {
			t.Errorf("expected value to be nil, got '%s'", r1)
		}
		func(cur realm) { store.Set(0, cur, []byte("hello")) }(cross2(cur))
		if r2 := store.Get(caller); string(r2) != "hello" {
			t.Errorf("expected value to be 'hello', got '%s'", r2)
		}
		func(cur realm) { store.Set(0, cur, []byte("world")) }(cross2(cur))
		if r3 := store.Get(caller); string(r3) != "world" {
			t.Errorf("expected value to be 'world', got '%s'", r3)
		}
	})
	t.Run("TestGet", func(cur realm, t *testing.T) {
		testing.SetRealm(testing.NewUserRealm(caller))
		store := New()
		func(cur realm) { store.Set(0, cur, []byte("hello")) }(cross2(cur))
		if r0 := store.Get(testutils.TestAddress("nil_user")); r0 != nil {
			t.Errorf("expected value to be nil, got '%s'", r0)
		}
		if r1 := store.Get(caller); string(r1) != "hello" {
			t.Errorf("expected value to be 'hello', got '%s'", r1)
		}
	})
}
