package json

import (
    "testing"
    "math"
)

type elTestCase struct {
    man    uint64
    exp10  int
    neg    bool
    want   float64
    ok     bool
}

var testCases = []elTestCase {
    {12345, 3, false, 12345e3, true},
    {12345, -3, false, 12345e-3, true},
}

func TestEiselLemire64(t *testing.T) {
    for _, tc := range testCases {
        got, ok := eiselLemire64(tc.man, tc.exp10, tc.neg)
        if ok != tc.ok || (ok && got != tc.want) {
            t.Errorf("eiselLemire64(%v, %v, %v) = %v, %v; want %v, %v", tc.man, tc.exp10, tc.neg, got, ok, tc.want, tc.ok)
        }
    }
}

func TestEiselLemire32(t *testing.T) {
    testCases := []elTestCase {
        {12345, 3, false, 12345e3, true},
        {12345, -3, false, 12345e-3, true},
    }

    for _, tc := range testCases {
        got, ok := eiselLemire32(tc.man, tc.exp10, tc.neg)
        want := float32(tc.want)
        if ok != tc.ok || (ok && got != want) {
            t.Errorf("eiselLemire32(%v, %v, %v) = %v, %v; want %v, %v", tc.man, tc.exp10, tc.neg, got, ok, want, tc.ok)
        }
    }
}