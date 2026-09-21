package gnoland

import (
	"strings"
	"testing"
)

// TestGnoSessionAccountValidateAllowPaths confirms that *GnoSessionAccount
// satisfies the auth handler's local allowPathsValidator interface and
// delegates correctly to the parser. The grammar itself is covered by
// TestParseAllowPathsEntry; this test is the wiring assertion.
func TestGnoSessionAccountValidateAllowPaths(t *testing.T) {
	t.Parallel()

	// Compile-time check: *GnoSessionAccount must implement the validator
	// interface that handleMsgCreateSession looks up.
	var _ interface{ ValidateAllowPaths([]string) error } = (*GnoSessionAccount)(nil)

	a := &GnoSessionAccount{}

	tests := []struct {
		name    string
		paths   []string
		wantErr string
	}{
		{name: "empty input rejected", paths: nil, wantErr: "AllowPaths is required"},
		{name: "wildcard accepted", paths: []string{"*"}},
		{name: "valid grammar ok", paths: []string{"vm/exec:gno.land/r/foo", "bank/send"}},
		{name: "bare route rejected", paths: []string{"bank"}, wantErr: "unknown route_type"},
		{name: "privilege escalation rejected", paths: []string{"auth/create_session"}, wantErr: "unknown route_type"},
		{name: "path on non-exec rejected", paths: []string{"bank/send:gno.land/r/foo"}, wantErr: "only vm/exec accepts a path"},
		{name: "wildcard with path rejected", paths: []string{"*:gno.land/r/foo"}, wantErr: "wildcard"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := a.ValidateAllowPaths(tc.paths)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
