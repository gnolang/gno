package gnoland

import (
	"strings"
	"testing"
)

func TestParseAllowPathsEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    allowPathsEntry
		wantErr string // substring match; "" means no error
	}{
		// Accepted shapes.
		{
			name:  "vm/exec without path",
			input: "vm/exec",
			want:  allowPathsEntry{Route: "vm", Type: "exec"},
		},
		{
			name:  "vm/exec with path",
			input: "vm/exec:gno.land/r/foo",
			want:  allowPathsEntry{Route: "vm", Type: "exec", Path: "gno.land/r/foo"},
		},
		{
			name:  "vm/run",
			input: "vm/run",
			want:  allowPathsEntry{Route: "vm", Type: "run"},
		},
		{
			name:  "bank/send",
			input: "bank/send",
			want:  allowPathsEntry{Route: "bank", Type: "send"},
		},
		{
			name:  "bank/multisend",
			input: "bank/multisend",
			want:  allowPathsEntry{Route: "bank", Type: "multisend"},
		},
		{
			name:  "path containing colons survives SplitN",
			input: "vm/exec:gno.land/r/foo:bar",
			want:  allowPathsEntry{Route: "vm", Type: "exec", Path: "gno.land/r/foo:bar"},
		},
		{
			name:  "deep realm path",
			input: "vm/exec:gno.land/r/jae/blog",
			want:  allowPathsEntry{Route: "vm", Type: "exec", Path: "gno.land/r/jae/blog"},
		},

		// Rejected: bare-route forms (future relaxation).
		{
			name:    "bare bank rejected",
			input:   "bank",
			wantErr: "unknown route_type",
		},
		{
			name:    "bare vm rejected",
			input:   "vm",
			wantErr: "unknown route_type",
		},

		// Privilege-escalation guard: not in whitelist → rejected.
		{name: "auth/create_session rejected", input: "auth/create_session", wantErr: "unknown route_type"},
		{name: "vm/add_package rejected", input: "vm/add_package", wantErr: "unknown route_type"},

		// Malformed shapes.
		{name: "empty string rejected", input: "", wantErr: "empty"},
		{name: "unknown type rejected", input: "bank/foo", wantErr: "unknown route_type"},
		{name: "case-sensitive rejected", input: "VM/EXEC", wantErr: "unknown route_type"},
		{name: "extra slash rejected", input: "vm/exec/extra", wantErr: "unknown route_type"},

		// Rejected: bad path suffix.
		{
			name:    "path on bank/send rejected",
			input:   "bank/send:gno.land/r/foo",
			wantErr: "only vm/exec accepts a path",
		},
		{
			name:    "path on vm/run rejected",
			input:   "vm/run:gno.land/r/foo",
			wantErr: "only vm/exec accepts a path",
		},
		{
			name:    "empty path after colon rejected",
			input:   "vm/exec:",
			wantErr: "non-empty path",
		},
		{
			name:    "trailing slash rejected",
			input:   "vm/exec:gno.land/r/foo/",
			wantErr: "trailing slash",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseAllowPathsEntry(tc.input)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("parseAllowPathsEntry(%q): expected error containing %q, got nil (got %+v)", tc.input, tc.wantErr, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("parseAllowPathsEntry(%q): error = %v, want substring %q", tc.input, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAllowPathsEntry(%q): unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseAllowPathsEntry(%q) = %+v, want %+v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseAllowPaths_Slice(t *testing.T) {
	t.Parallel()

	t.Run("empty input rejected", func(t *testing.T) {
		t.Parallel()
		_, err := parseAllowPaths(nil)
		if err == nil {
			t.Fatal("expected error for empty input, got nil")
		}
		if !strings.Contains(err.Error(), "AllowPaths is required") {
			t.Errorf("expected 'AllowPaths is required' error, got %v", err)
		}
	})

	t.Run("wildcard accepted", func(t *testing.T) {
		t.Parallel()
		got, err := parseAllowPaths([]string{"*"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || !got[0].Wildcard {
			t.Errorf("expected wildcard entry, got %+v", got)
		}
	})

	t.Run("wildcard with path rejected", func(t *testing.T) {
		t.Parallel()
		_, err := parseAllowPaths([]string{"*:gno.land/r/foo"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "wildcard") {
			t.Errorf("expected wildcard error, got %v", err)
		}
	})

	t.Run("multiple valid entries", func(t *testing.T) {
		t.Parallel()
		input := []string{"vm/exec:gno.land/r/foo", "bank/send", "vm/run"}
		got, err := parseAllowPaths(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(got))
		}
		if got[0].Path != "gno.land/r/foo" || got[1].Type != "send" || got[2].Type != "run" {
			t.Errorf("unexpected parse: %+v", got)
		}
	})

	t.Run("first invalid entry is reported with index", func(t *testing.T) {
		t.Parallel()
		input := []string{"vm/exec", "bank/foo", "vm/run"}
		_, err := parseAllowPaths(input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "allow_paths[1]") {
			t.Errorf("expected error to cite allow_paths[1], got %v", err)
		}
	})

	t.Run("duplicates and shadowing accepted (idempotent)", func(t *testing.T) {
		t.Parallel()
		// Pinned so a future "tightening" doesn't silently break callers that
		// wrote both bare and path-bearing forms, or duplicates.
		input := []string{"vm/exec", "vm/exec:gno.land/r/foo", "vm/exec:gno.land/r/foo"}
		got, err := parseAllowPaths(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("expected 3 entries preserved, got %d", len(got))
		}
	})
}
