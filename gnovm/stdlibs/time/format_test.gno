// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time_test

import (
	"testing"
	"time"
)

// Custom code: gno's time package shipped with no tests at all. These cover the
// parsing and comparison behaviour that diverged from Go, so the fixes cannot
// silently regress. They are adapted from Go's time/format_test.go and
// time/time_test.go rather than copied wholesale, because Go's suites depend on
// zoneinfo files, monotonic clocks and testing/quick.

func TestParseTimeZoneOffsetRange(t *testing.T) {
	// Go rejects offsets whose components are out of range, and rejects a
	// signed component inside the offset (getnum, not atoi).
	bad := []struct{ layout, value string }{
		{"-0700", "+-500"},  // sign inside the hour field
		{"-0700", "+9900"},  // hour out of range
		{"-07:00", "+05:99"}, // minute out of range
		{"-07:00", "-00:99"}, // minute out of range, negative offset
	}
	for _, tc := range bad {
		if got, err := time.Parse(tc.layout, tc.value); err == nil {
			t.Errorf("time.Parse(%q, %q) = %v, want error", tc.layout, tc.value, got)
		}
	}

	good := []struct {
		layout, value string
		wantOffset    int
	}{
		{"-0700", "+0530", 5*3600 + 30*60},
		{"-0700", "-0800", -8 * 3600},
		{"-07:00", "+05:45", 5*3600 + 45*60},
		{"-0700", "+2400", 24 * 3600}, // 24h is deliberately allowed
	}
	for _, tc := range good {
		tm, err := time.Parse(tc.layout, tc.value)
		if err != nil {
			t.Errorf("time.Parse(%q, %q) failed: %v", tc.layout, tc.value, err)
			continue
		}
		if _, off := tm.Zone(); off != tc.wantOffset {
			t.Errorf("time.Parse(%q, %q) offset = %d, want %d", tc.layout, tc.value, off, tc.wantOffset)
		}
	}
}

func TestParseMissingSecondsBeforeFraction(t *testing.T) {
	// A fractional second must not paper over a missing seconds field: the
	// getnum failure has to survive rather than being overwritten by
	// parseNanoseconds.
	for _, value := range []string{"03:04:.99", "03:04:.5", "03:04:,25"} {
		if got, err := time.Parse("15:04:05", value); err == nil {
			t.Errorf("time.Parse(15:04:05, %q) = %v, want error", value, got)
		}
	}
	// The valid forms must keep working.
	tm, err := time.Parse("15:04:05", "03:04:05")
	if err != nil {
		t.Fatalf("time.Parse of a valid time failed: %v", err)
	}
	if tm.Hour() != 3 || tm.Minute() != 4 || tm.Second() != 5 {
		t.Errorf("time.Parse = %02d:%02d:%02d, want 03:04:05", tm.Hour(), tm.Minute(), tm.Second())
	}
	// A layout that does carry a fractional second still accepts one.
	tm, err = time.Parse("15:04:05.000", "03:04:05.250")
	if err != nil {
		t.Fatalf("time.Parse with fractional layout failed: %v", err)
	}
	if tm.Nanosecond() != 250000000 {
		t.Errorf("Nanosecond = %d, want 250000000", tm.Nanosecond())
	}
}

func TestParseFractionalSecondsBeyondNine(t *testing.T) {
	// More than nine fractional digits must not produce a spurious
	// "extra text" error; the digits past nanosecond precision are consumed.
	if _, err := time.Parse("15:04:05", "03:04:05.1234567890"); err != nil {
		t.Errorf("time.Parse with 10 fractional digits failed: %v", err)
	}
}

func TestCompare(t *testing.T) {
	a := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	b := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	c := time.Date(2020, time.January, 1, 0, 0, 0, 5, time.UTC)

	if got := a.Compare(b); got != -1 {
		t.Errorf("a.Compare(b) = %d, want -1", got)
	}
	if got := b.Compare(a); got != +1 {
		t.Errorf("b.Compare(a) = %d, want +1", got)
	}
	if got := a.Compare(a); got != 0 {
		t.Errorf("a.Compare(a) = %d, want 0", got)
	}
	// Same second, different nanoseconds.
	if got := a.Compare(c); got != -1 {
		t.Errorf("a.Compare(c) = %d, want -1", got)
	}
	// Compare must agree with Before/After/Equal.
	for _, pair := range [][2]time.Time{{a, b}, {b, a}, {a, a}, {a, c}} {
		x, y := pair[0], pair[1]
		got := x.Compare(y)
		switch {
		case got < 0 && !x.Before(y):
			t.Errorf("Compare < 0 but Before is false")
		case got > 0 && !x.After(y):
			t.Errorf("Compare > 0 but After is false")
		case got == 0 && !x.Equal(y):
			t.Errorf("Compare == 0 but Equal is false")
		}
	}
}

func TestMarshalRejectsUnrepresentableZone(t *testing.T) {
	// RFC 3339 cannot express an offset outside [-23:59, +23:59].
	tm := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.FixedZone("bad", 24*3600))
	if _, err := tm.MarshalJSON(); err == nil {
		t.Error("MarshalJSON should reject a 24h zone offset")
	}
	// A representable zone still marshals.
	ok := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.FixedZone("ok", 5*3600))
	b, err := ok.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON of a valid zone failed: %v", err)
	}
	if string(b) != `"2020-01-01T00:00:00+05:00"` {
		t.Errorf("MarshalJSON = %s", b)
	}
}

func TestRoundTripRFC3339(t *testing.T) {
	orig := time.Date(2024, time.March, 15, 12, 34, 56, 789000000, time.UTC)
	b, err := orig.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var back time.Time
	if err := back.UnmarshalJSON(b); err != nil {
		t.Fatal(err)
	}
	if !back.Equal(orig) {
		t.Errorf("round trip: got %v want %v", back, orig)
	}
}
