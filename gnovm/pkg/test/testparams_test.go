package test

import (
	"reflect"
	"testing"
)

func TestTestParams_MissingKeyLeavesPtrAtZero(t *testing.T) {
	tp := newTestParams()

	// Pre-set ptrs to non-zero so we can detect mutation.
	b := true
	var i int64 = 7
	var s = "preset"
	var ss = []string{"preset"}

	if tp.GetBool("missing", &b) {
		t.Fatal("GetBool on absent key must return false")
	}
	if tp.GetInt64("missing", &i) {
		t.Fatal("GetInt64 on absent key must return false")
	}
	if tp.GetString("missing", &s) {
		t.Fatal("GetString on absent key must return false")
	}
	if tp.GetStrings("missing", &ss) {
		t.Fatal("GetStrings on absent key must return false")
	}

	// Get on absent key must NOT mutate the destination.
	if !b || i != 7 || s != "preset" || !reflect.DeepEqual(ss, []string{"preset"}) {
		t.Fatalf("Get on absent key mutated ptr: b=%v i=%d s=%q ss=%v", b, i, s, ss)
	}
}

func TestTestParams_DistinguishesSetToZeroFromUnset(t *testing.T) {
	// The whole point of returning bool is to distinguish "key was
	// set to the zero value" from "key was never set." Verify for
	// each zero-able type.
	tp := newTestParams()
	tp.SetBool("set-false", false)
	tp.SetInt64("set-zero", 0)
	tp.SetString("set-empty", "")
	tp.SetStrings("set-empty-list", []string{})
	tp.SetBytes("set-empty-bytes", []byte{})

	cases := []struct {
		name string
		get  func() bool
	}{
		{"set-false bool", func() bool { var v bool; return tp.GetBool("set-false", &v) }},
		{"set-zero int64", func() bool { var v int64; return tp.GetInt64("set-zero", &v) }},
		{"set-empty string", func() bool { var v string; return tp.GetString("set-empty", &v) }},
		{"set-empty []string", func() bool { var v []string; return tp.GetStrings("set-empty-list", &v) }},
		{"set-empty []byte", func() bool { var v []byte; return tp.GetBytes("set-empty-bytes", &v) }},
	}
	for _, c := range cases {
		if !c.get() {
			t.Errorf("%s: found should be true (key was set, value happens to be zero)", c.name)
		}
	}

	// And the other direction: never-set keys.
	cases2 := []struct {
		name string
		get  func() bool
	}{
		{"unset bool", func() bool { var v bool; return tp.GetBool("nope", &v) }},
		{"unset int64", func() bool { var v int64; return tp.GetInt64("nope", &v) }},
		{"unset string", func() bool { var v string; return tp.GetString("nope", &v) }},
		{"unset []string", func() bool { var v []string; return tp.GetStrings("nope", &v) }},
		{"unset []byte", func() bool { var v []byte; return tp.GetBytes("nope", &v) }},
	}
	for _, c := range cases2 {
		if c.get() {
			t.Errorf("%s: found should be false", c.name)
		}
	}
}

func TestTestParams_TypeMismatchFailsSafe(t *testing.T) {
	tp := newTestParams()
	tp.SetBool("k", true)

	var got []string
	if tp.GetStrings("k", &got) {
		t.Fatal("type-mismatched Get must return false")
	}
	if got != nil {
		t.Fatalf("type-mismatched Get must leave ptr at zero, got %v", got)
	}
}

func TestTestParams_UpdateStringsDedupesOnAdd(t *testing.T) {
	tp := newTestParams()
	tp.SetStrings("k", []string{"a", "b"})

	tp.UpdateStrings("k", []string{"b", "c", "c"}, true)

	var got []string
	tp.GetStrings("k", &got)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTestParams_UpdateStringsRemove(t *testing.T) {
	tp := newTestParams()
	tp.SetStrings("k", []string{"a", "b", "c"})

	tp.UpdateStrings("k", []string{"b", "missing"}, false)

	var got []string
	tp.GetStrings("k", &got)
	want := []string{"a", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTestParams_UpdateStringsAddOnEmpty(t *testing.T) {
	tp := newTestParams()

	tp.UpdateStrings("k", []string{"a", "a", "b"}, true)

	var got []string
	tp.GetStrings("k", &got)
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
