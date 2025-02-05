package testing_test

import (
	"testing"
	"testing/fuzzing"
	"unicode/utf8"
)

/// TODO: Once coverage is fully implemented, test whether the fuzzer can detect
//       issues such as HTTP request failures or compilation errors.

// FuzzEdgeCase tests whether the fuzzer can effectively generate edge cases.
// It evaluates the ability to produce edge inputs in a structured manner.
func FuzzEdgeCase(f *testing.F) {
	f.Add("apple hello", int(42131231230))
	f.Add("rainy day", int(98401132231331))
	f.Add("winter comes", int(12349123123))
	f.Fuzz(func(t *testing.T, orig ...interface{}) {
		v, ok := orig[0].(string)
		if !ok {
			panic("dont match")
		}
		i, ok2 := orig[1].(int)
		if !ok2 {
			panic("dont match")
		}
		rev := fuzzing.Reverse(v)
		doubleRev := fuzzing.Reverse(rev)
		if v != doubleRev && i > 300 && i < 500 {
			t.Errorf("Before: %q, after: %q", orig, doubleRev)
		}
		if utf8.ValidString(v) && !utf8.ValidString(rev) && i > 300 && i < 1000 {
			t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
		}
	})
}

// FuzzSymbolicPath tests whether the fuzzer can explore symbolic execution paths.
// It assesses the ability to traverse deep path depths effectively.
func FuzzSymbolicPath(f *testing.F) {
	f.Add("")
	f.Fuzz(func(t *testing.T, orig ...interface{}) {
		s, ok := orig[0].(string)
		if !ok {
			panic("dont match")
		}
		if len(s) > 0 && s[0] == 'b' {
			if len(s) > 1 && s[1] == 'a' {
				if len(s) > 2 && s[2] == 'd' {
					if len(s) > 3 && s[3] == '!' {
						panic("crash triggered")
					}
				}
			}
		}
	})
}
