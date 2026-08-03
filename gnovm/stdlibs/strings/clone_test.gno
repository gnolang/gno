// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"strings"
	"testing"
)

func TestClone(t *testing.T) {
	var cloneTests = []string{
		"",
		strings.Clone(""),
		strings.Repeat("a", 42)[:0],
		"short",
		strings.Repeat("a", 42),
	}
	for _, input := range cloneTests {
		clone := strings.Clone(input)
		if clone != input {
			t.Errorf("Clone(%q) = %q; want %q", input, clone, input)
		}
	}
}

// Subset of upstream ContainsRuneTests; the full table lives in
// strings_test.go and will arrive with PR #5749.
var containsFuncTests = []struct {
	str      string
	r        rune
	expected bool
}{
	{"", 'a', false},
	{"a", 'a', true},
	{"aaa", 'a', true},
	{"abc", 'y', false},
	{"x\xc2\x4cz", 'L', true},
	{"x\xc2\x4cz", '䰀', false},
}

func TestContainsFunc(t *testing.T) {
	for _, ct := range containsFuncTests {
		if strings.ContainsFunc(ct.str, func(r rune) bool {
			return ct.r == r
		}) != ct.expected {
			t.Errorf("ContainsFunc(%q, func(%q)) = %v, want %v",
				ct.str, ct.r, !ct.expected, ct.expected)
		}
	}
}
