// Whitebox tests for filterMatch.verify methods:
// These tests exercise unreachable methods simpleMatch.verify and alternationMatch.verify
// solely to not negatively impact code coverage metrics. These methods are not invoked by production code
// and are included only to prevent negative impact on test coverage.

package test

import (
	"testing"
)

// matchString is the function used by verify methods to validate patterns.
// It compiles the pattern as a regular expression.

func TestSimpleMatchVerify_ValidPatterns(t *testing.T) {
	cases := []string{
		"abc",           // plain literal
		"a.c",           // dot meta-character
		"^test$",        // anchors
		"[a-z]+",        // character class
		"hello\\ world", // escaped space (\)
		"\u2000\u202f",  // two unicode space characters
		"\u0007",        // non-printable (bell)
	}
	for _, pat := range cases {
		sm := simpleMatch{pat}
		if err := sm.verify(pat, matchString); err != nil {
			t.Errorf("simpleMatch.verify(%q) returned unexpected error: %v", pat, err)
		}
	}
}

func TestSimpleMatchVerify_InvalidPatterns(t *testing.T) {
	cases := []string{
		"(",     // unmatched parenthesis
		"[a-",   // unterminated character class
		"**bad", // invalid repetition
	}
	for _, pat := range cases {
		sm := simpleMatch{pat}
		err := sm.verify(pat, matchString)
		if err == nil {
			t.Errorf("simpleMatch.verify(%q) expected error, got nil", pat)
		}
	}
}

func TestAlternationMatchVerify_ValidPatterns(t *testing.T) {
	cases := []string{
		"a|b",        // simple alternation
		"foo|bar|baz",// multiple alternations
		"x[0-9]|y[0-9]", // alternations with regex
	}
	for _, pat := range cases {
		am := alternationMatch{simpleMatch{pat}}
		if err := am.verify(pat, matchString); err != nil {
			t.Errorf("alternationMatch.verify(%q) returned unexpected error: %v", pat, err)
		}
	}
}

func TestAlternationMatchVerify_InvalidPatterns(t *testing.T) {
	cases := []string{
		"**bad/**things",  // starts with |
		"**one_bad",  // trailing |
		"/??fail/",  // empty alternative
	}
	for _, pat := range cases {
		// splitRegexp will parse these into alternations, then verify each
		fm := splitRegexp(pat)
		err := fm.verify(pat, matchString)
		if err == nil {
			t.Errorf("filterMatch.verify(%q) expected error, got nil", pat)
		}
	}
}
