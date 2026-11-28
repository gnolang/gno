package gnoweb

import (
	"net/url"
	"testing"
)

// Ownership is stated as a rule, so it can be checked without standing up a
// handler — and without depending on which dispatch block runs first, which
// is the whole reason the predicate exists.
func TestOwnsSearchURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args string
		want bool
	}{
		{"plain search", "search&q=foo", true},
		{"search json", "search&q=foo&json", true},
		// feature/state's own filtered view carries both keys.
		{"state with a search filter", "state&search=foo", false},
		{"state alone", "state", false},
		{"neither", "source&file=x.gno", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wq, err := url.ParseQuery(tt.args)
			if err != nil {
				t.Fatalf("ParseQuery(%q): %v", tt.args, err)
			}
			if got := ownsSearchURL(wq); got != tt.want {
				t.Fatalf("ownsSearchURL(%q) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
