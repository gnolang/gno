package keyscli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// nopCloser wraps a writer to satisfy io.WriteCloser.
type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func newTestIO() (commands.IO, *bytes.Buffer) {
	var errBuf bytes.Buffer
	cio := commands.NewDefaultIO()
	cio.SetErr(nopCloser{&errBuf})
	cio.SetOut(nopCloser{io.Discard})
	return cio, &errBuf
}

func TestEvalVersionGap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		basePath   string
		version    int
		result     *vm.LatestVersionResult
		force      bool
		wantErr    string // substring of error, empty means no error
		wantStderr string // substring of stderr warning
	}{
		{
			name:     "sequential_no_warning",
			basePath: "gno.land/p/demo/avl",
			version:  2,
			result:   &vm.LatestVersionResult{Latest: "v1"},
			wantErr:  "",
		},
		{
			name:       "small_gap_warns",
			basePath:   "gno.land/p/demo/avl",
			version:    3,
			result:     &vm.LatestVersionResult{Latest: "v1"},
			wantErr:    "",
			wantStderr: "Warning:",
		},
		{
			name:     "large_gap_blocks",
			basePath: "gno.land/p/demo/avl",
			version:  10,
			result:   &vm.LatestVersionResult{Latest: "v1"},
			wantErr:  "version gap too large",
		},
		{
			name:     "large_gap_force_overrides",
			basePath: "gno.land/p/demo/avl",
			version:  10,
			result:   &vm.LatestVersionResult{Latest: "v1"},
			force:    true,
			wantErr:  "",
		},
		{
			name:       "no_versions_on_chain",
			basePath:   "gno.land/p/demo/avl",
			version:    2,
			result:     nil,
			wantErr:    "",
			wantStderr: "no previous versions",
		},
		{
			name:     "no_versions_large_gap_blocks",
			basePath: "gno.land/p/demo/avl",
			version:  10,
			result:   nil,
			wantErr:  "version gap too large",
		},
		{
			name:     "no_versions_large_gap_force",
			basePath: "gno.land/p/demo/avl",
			version:  10,
			result:   nil,
			force:    true,
			wantErr:  "",
		},
		{
			name:       "short_name_in_warning",
			basePath:   "gno.land/p/demo/versiontest",
			version:    3,
			result:     &vm.LatestVersionResult{Latest: "v1"},
			wantStderr: "versiontest/v3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cio, errBuf := newTestIO()

			err := evalVersionGap(tt.basePath, tt.version, tt.result, tt.force, cio)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
			}

			stderr := errBuf.String()
			if tt.wantStderr != "" && !strings.Contains(stderr, tt.wantStderr) {
				t.Fatalf("stderr %q does not contain %q", stderr, tt.wantStderr)
			}
		})
	}
}
