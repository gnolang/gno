package loci

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestLociStore(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test/test"))

	t.Run("TestSet", func(t *testing.T) {
		t.Parallel()
		store := New()
		u1 := testutils.TestAddress("u1")

		m1 := []byte("hello")
		m2 := []byte("world")
		testing.SetOriginCaller(u1)

		// Ensure that the value is nil before setting it.
		r1 := store.Get(u1)
		if r1 != nil {
			t.Errorf("expected value to be nil, got '%s'", r1)
		}
		store.Set(m1)
		// Ensure that the value is correct after setting it.
		r2 := store.Get(u1)
		if string(r2) != "hello" {
			t.Errorf("expected value to be 'hello', got '%s'", r2)
		}
		store.Set(m2)
		// Ensure that the value is correct after overwriting it.
		r3 := store.Get(u1)
		if string(r3) != "world" {
			t.Errorf("expected value to be 'world', got '%s'", r3)
		}
	})
	t.Run("TestGet", func(t *testing.T) {
		t.Parallel()
		store := New()
		u1 := testutils.TestAddress("u1")
		u2 := testutils.TestAddress("u2")
		u3 := testutils.TestAddress("u3")
		u4 := testutils.TestAddress("u4")

		m1 := []byte("hello")
		m2 := []byte("world")
		m3 := []byte("goodbye")

		testing.SetOriginCaller(u1)
		store.Set(m1)
		testing.SetOriginCaller(u2)
		store.Set(m2)
		testing.SetOriginCaller(u3)
		store.Set(m3)

		// Ensure that the value is correct after setting it.
		r0 := store.Get(u4)
		if r0 != nil {
			t.Errorf("expected value to be nil, got '%s'", r0)
		}
		// Ensure that the value is correct after setting it.
		r1 := store.Get(u1)
		if string(r1) != "hello" {
			t.Errorf("expected value to be 'hello', got '%s'", r1)
		}
		// Ensure that the value is correct after setting it.
		r2 := store.Get(u2)
		if string(r2) != "world" {
			t.Errorf("expected value to be 'world', got '%s'", r2)
		}
		// Ensure that the value is correct after setting it.
		r3 := store.Get(u3)
		if string(r3) != "goodbye" {
			t.Errorf("expected value to be 'goodbye', got '%s'", r3)
		}
	})
}
