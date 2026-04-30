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

	tp.GetBool("missing", &b)
	tp.GetInt64("missing", &i)
	tp.GetString("missing", &s)
	tp.GetStrings("missing", &ss)

	// Get on absent key must NOT mutate the destination.
	if !b || i != 7 || s != "preset" || !reflect.DeepEqual(ss, []string{"preset"}) {
		t.Fatalf("Get on absent key mutated ptr: b=%v i=%d s=%q ss=%v", b, i, s, ss)
	}
}

func TestTestParams_TypeMismatchFailsSafe(t *testing.T) {
	tp := newTestParams()
	tp.SetBool("k", true)

	var got []string
	tp.GetStrings("k", &got)
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
