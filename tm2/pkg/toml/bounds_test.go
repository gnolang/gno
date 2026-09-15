// This file is not from upstream go-toml: it is the regression suite for the
// resource bounds this fork adds (maxNestingDepth, maxKeyDepth, and the O(1)
// duplicate-table check). See README.md.

package toml

import (
	"fmt"
	"strings"
	"testing"
)

func TestBoundsNestingDepth(t *testing.T) {
	t.Parallel()

	// Closers are not needed to descend: parseRvalue/parseArray recurse on the
	// opener alone, so an unclosed run is the cheapest way to buy depth. Both
	// forms must be bounded.
	for _, tc := range []struct {
		name    string
		body    string
		wantErr string
	}{
		{"at limit", "a = " + strings.Repeat("[", maxNestingDepth) + strings.Repeat("]", maxNestingDepth), ""},
		{"over limit, balanced", "a = " + strings.Repeat("[", maxNestingDepth+1) + strings.Repeat("]", maxNestingDepth+1), "nesting depth exceeds limit"},
		{"over limit, unclosed", "a = " + strings.Repeat("[", maxNestingDepth+1), "nesting depth exceeds limit"},
		{"inline tables", "a = " + strings.Repeat("{b=", maxNestingDepth+1), "nesting depth exceeds limit"},
		// The payload that fatally overflows the goroutine stack upstream: a
		// stack overflow is not a panic, so no recover() can catch it and the
		// whole process dies. ~400k openers is under 400KB.
		{"stack overflow payload", "a = " + strings.Repeat("[", 420_000), "nesting depth exceeds limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Load(tc.body)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestBoundsKeyDepth(t *testing.T) {
	t.Parallel()

	dotted := func(n int) string { return strings.Repeat("a.", n-1) + "z" }
	// parseKey appends a key group per quoted run and does not require a '.'
	// between segments, so ["a""b""c"] is the path a.b.c. Counting the parsed
	// path rather than the bytes is what makes both forms cost the same.
	quoted := func(n int) string { return strings.Repeat(`""`, n) }

	for _, tc := range []struct {
		name    string
		body    string
		wantErr string
	}{
		{"dotted at limit", "[" + dotted(maxKeyDepth) + "]\nk = 1", ""},
		{"dotted over limit", "[" + dotted(maxKeyDepth+1) + "]\nk = 1", "key path depth"},
		{"quoted at limit", "[" + quoted(maxKeyDepth) + "]\nk = 1", ""},
		{"quoted over limit", "[" + quoted(maxKeyDepth+1) + "]\nk = 1", "key path depth"},
		{"assignment key over limit", dotted(maxKeyDepth+1) + " = 1", "key path depth"},
		{"table array over limit", "[[" + dotted(maxKeyDepth+1) + "]]", "key path depth"},
		// A quoted segment may span newlines (lexInsideTableKey copies bytes
		// through to parseKey until ']', and the quoted-key branch consumes to
		// the closing quote), so a key path can hide its depth from any
		// per-line byte count. Counting the parsed path sees it regardless.
		{"key path spanning lines", "[\"\"" + strings.Repeat(".\"\n\"", maxKeyDepth+1) + "]", "key path depth"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Load(tc.body)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

// Duplicate tables are still rejected after the O(n^2) scan became an O(1) set
// lookup, including across the table-array pruning that deletes from the set.
func TestBoundsDuplicateTableStillDetected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"duplicate table", "[a]\nk = 1\n[a]\nj = 2", true},
		{"distinct tables", "[a]\nk = 1\n[b]\nj = 2", false},
		{"table array then same table", "[[a]]\nk = 1\n[a]\nj = 2", true},
		{"repeated table array is legal", "[[a]]\nk = 1\n[[a]]\nk = 2", false},
		{"subtable of table array", "[[a]]\n[a.b]\nk = 1\n[[a]]\n[a.b]\nk = 2", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Load(tc.body)
			if got := err != nil; got != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

// The table count is linear, not quadratic: 20k distinct tables used to make
// parseGroup rescan a growing slice once per table.
func TestBoundsTableCountIsLinear(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	for i := range 20_000 {
		fmt.Fprintf(&b, "[t%06d]\nk = 1\n", i)
	}
	if _, err := Load(b.String()); err != nil {
		t.Fatal(err)
	}
}

// Ordinary documents are unaffected by any of the bounds.
func TestBoundsAcceptOrdinaryDocuments(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		"module = \"gno.land/r/demo/foo\"\ngno = \"0.9\"\n",
		"# don't bump this\nmodule = \"gno.land/r/demo/foo\"\ngno = \"0.9\"\n",
		"module = \"gno.land/r/f\"\n[[replace]]\nold = \"gno.land/p/demo/avl\"\nnew = \"..\\\\..\\\\p\\\\avl\"\n",
		"a = [1, 2, [3, [4]]]\n[t]\nb = {c = 1, d = {e = 2}}\n",
		"a.b.c = 1\n[x.y.z]\nk = 2\n",
	} {
		if _, err := Load(body); err != nil {
			t.Errorf("rejected an ordinary document: %v\n%s", err, body)
		}
	}
}

// Marshal's output must decode back to the value it was given, for every rune.
//
// PackageContentHash (gno.land/pkg/sdk/vm) hashes gnomod.toml after a
// parse-and-re-encode, against a stored copy the keeper had already parsed and
// re-encoded once. The two agree only if encoding is a fixpoint, so a rune that
// survives one round trip but not two is an approval nobody can match.
//
// Upstream's control-character escape got this wrong in both directions:
// `uint16(rr) < 0x001F` excluded U+001F, which was then written raw into a
// basic string the lexer refuses; and the conversion truncated, so a rune whose
// low 16 bits fall under 0x1F was emitted as the escape for those bits alone.
func TestBoundsMarshalRoundTripsEveryRune(t *testing.T) {
	t.Parallel()

	type doc struct {
		V string `toml:"v"`
	}

	check := func(t *testing.T, in string) {
		t.Helper()
		first, err := Marshal(doc{V: in})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var back doc
		if err := Unmarshal(first, &back); err != nil {
			t.Fatalf("the encoder produced a document the decoder refuses: %v\n%q", err, first)
		}
		if back.V != in {
			t.Fatalf("value changed across a round trip: %q -> %q", in, back.V)
		}
		second, err := Marshal(back)
		if err != nil {
			t.Fatalf("re-marshal: %v", err)
		}
		if string(first) != string(second) {
			t.Fatalf("encoding is not a fixpoint:\n first: %q\nsecond: %q", first, second)
		}
	}

	t.Run("every control character", func(t *testing.T) {
		t.Parallel()
		for r := rune(0); r <= 0x1F; r++ {
			check(t, "x"+string(r)+"y")
		}
	})

	t.Run("astral runes with low bits under 0x1F", func(t *testing.T) {
		t.Parallel()
		for _, r := range []rune{0x1000A, 0x10000, 0x1001E, 0x2000D, 0x10FFFF} {
			check(t, "x"+string(r)+"y")
		}
	})
}
