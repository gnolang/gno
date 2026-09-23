package omnisearch

import (
	"errors"
	"strings"
	"testing"
)

func TestParseQuerySplitsQualifiersFromText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		raw         string
		wantText    string
		wantFilters []Filter
	}{
		{
			name:     "plain text",
			raw:      "boards",
			wantText: "boards",
		},
		{
			name:        "one qualifier",
			raw:         "func:Vote",
			wantFilters: []Filter{{Key: "func", Value: "Vote"}},
		},
		{
			name:        "qualifier plus text",
			raw:         "author:alice boards",
			wantText:    "boards",
			wantFilters: []Filter{{Key: "author", Value: "alice"}},
		},
		{
			name:        "several qualifiers",
			raw:         "content:Vote author:alice is:realm",
			wantFilters: []Filter{{Key: "content", Value: "Vote"}, {Key: "author", Value: "alice"}, {Key: "is", Value: "realm"}},
		},
		{
			// A value may legitimately contain a colon: a package path, an
			// object id. Only the first one separates.
			name:        "value keeps its own colons",
			raw:         "in:gno.land/r/demo/boards:p/foo",
			wantFilters: []Filter{{Key: "in", Value: "gno.land/r/demo/boards:p/foo"}},
		},
		{
			name:        "key is case insensitive",
			raw:         "FUNC:Vote",
			wantFilters: []Filter{{Key: "func", Value: "Vote"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q, err := ParseQuery(tt.raw)
			if err != nil {
				t.Fatalf("ParseQuery(%q): %v", tt.raw, err)
			}
			if q.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", q.Text, tt.wantText)
			}
			if len(q.Filters) != len(tt.wantFilters) {
				t.Fatalf("Filters = %+v, want %+v", q.Filters, tt.wantFilters)
			}
			for i, f := range tt.wantFilters {
				if q.Filters[i] != f {
					t.Errorf("Filters[%d] = %+v, want %+v", i, q.Filters[i], f)
				}
			}
		})
	}
}

func TestParseQueryRejectsHostileInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want error
	}{
		{"over length", strings.Repeat("a", MaxQueryLen+1), ErrQueryTooLong},
		{"carriage return", "func:Vote\rSet-Cookie: x", ErrQueryInvalid},
		{"newline", "func:Vote\nx", ErrQueryInvalid},
		{"too many qualifiers", strings.Repeat("a:b ", MaxFilters+1), ErrTooManyFilters},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseQuery(tt.raw); !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestHasBareWord(t *testing.T) {
	t.Parallel()

	q, err := ParseQuery("imports")
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	if !q.HasBareWord("imports") {
		t.Error("HasBareWord(imports) = false, want true")
	}
	if q.HasBareWord("activity") {
		t.Error("HasBareWord(activity) = true, want false")
	}
}

func TestNormalizePkgPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"/r/demo/boards", "/r/demo/boards"},
		{"r/demo/boards", "/r/demo/boards"},
		{"gno.land/r/demo/boards", "/r/demo/boards"},
		{"gno.land/p/demo/avl", "/p/demo/avl"},
		// Anything that is not a realm or package path is refused rather than
		// passed to the chain as a package.
		{"/u/alice", ""},
		{"../../etc/passwd", ""},
		{"", ""},
		// `in:` is reader-supplied and becomes both a chain query and an href.
		{`/r/demo/a" onmouseover=alert(1) x="`, ""},
		{"/r/demo/a<script>", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := normalizePkgPath(tt.in, "gno.land"); got != tt.want {
				t.Fatalf("normalizePkgPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// Hrefs are built from chain-controlled strings, so anything that is not a
// plain path or namespace must produce no link at all.
func TestHrefsRefuseHostileSegments(t *testing.T) {
	t.Parallel()

	hostile := []string{
		"/u/x", "javascript:alert(1)", "/etc/passwd", "",
		// Deployment is permissionless, so a package path is attacker-supplied.
		// None of these may become a link.
		`/r/demo/a" onmouseover=alert(1) x="`,
		"/r/demo/a<script>",
		"/r/demo/a b",
		"/r/demo/a/../../etc",
		"/r/demo/a%2e%2e",
		"/r/demo/a?x=1",
		"/r/demo/a#frag",
	}
	for _, in := range hostile {
		if got := safePathHref(in); got != "" {
			t.Errorf("safePathHref(%q) = %q, want empty", in, got)
		}
	}
	if got := safePathHref("/r/demo/boards"); got != "/r/demo/boards" {
		t.Errorf("safePathHref(/r/demo/boards) = %q, want it to link", got)
	}
	for _, in := range []string{"a/b", "x?y", "x#y", "x$y", "", "a:b"} {
		if got := safeUserHref(in); got != "" {
			t.Errorf("safeUserHref(%q) = %q, want empty", in, got)
		}
	}
	if got := safeUserHref("alice"); got != "/u/alice" {
		t.Errorf("safeUserHref(alice) = %q, want /u/alice", got)
	}
}

// The omnibar is prefilled with the path the reader is on, so a path with
// arguments must never parse as a qualifier — otherwise pressing Enter
// searches for a qualifier that does not exist instead of navigating.
func TestPathsAndURLsAreNotQualifiers(t *testing.T) {
	t.Parallel()

	tests := []string{
		"/r/gnoland/pages:p/about",
		"/r/demo/boards:thread/1",
		"gno.land/r/gnoland/pages:p/about",
		"https://gno.land/r/demo/boards",
		"http://localhost:8888/r/demo",
		// Deliberately not here: a bare `localhost:8888` has the exact shape
		// of a qualifier, and reporting "localhost is not a qualifier" tells
		// the reader more than silently searching for it would.
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			q, err := ParseQuery(raw)
			if err != nil {
				t.Fatalf("ParseQuery(%q): %v", raw, err)
			}
			if len(q.Filters) != 0 {
				t.Fatalf("Filters = %+v, want none — %q is a path or a URL", q.Filters, raw)
			}
			if q.Text != raw {
				t.Fatalf("Text = %q, want the whole input %q", q.Text, raw)
			}
		})
	}
}

// …while a real qualifier still parses as one.
func TestBareWordKeysAreStillQualifiers(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"func:Vote", "author:alice", "in:/r/demo/boards", "content:foo"} {
		q, err := ParseQuery(raw)
		if err != nil {
			t.Fatalf("ParseQuery(%q): %v", raw, err)
		}
		if len(q.Filters) != 1 {
			t.Fatalf("%q: Filters = %+v, want one", raw, q.Filters)
		}
	}
}
