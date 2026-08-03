package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindTestScripts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		layout      []string // files to create, slash-separated relative paths
		roots       []string // roots to scan, relative to the temp dir; default "."
		want        []string
		errContains string
	}{
		{
			name:   "nested scripts and non-scripts",
			layout: []string{"r/demo/foo/foo.gno", "r/demo/foo/foo.txtar", "r/demo/bar/bar.txtar", "README.md"},
			want:   []string{"r/demo/bar/bar.txtar", "r/demo/foo/foo.txtar"},
		},
		{
			name:   "hidden directories are skipped",
			layout: []string{".git/hooks/nope.txtar", "r/ok.txtar"},
			want:   []string{"r/ok.txtar"},
		},
		{
			name:   "no scripts",
			layout: []string{"r/demo/foo/foo.gno"},
			want:   nil,
		},
		{
			name:        "duplicate base names",
			layout:      []string{"r/a/render.txtar", "r/b/render.txtar"},
			errContains: `duplicate testscript base name "render.txtar"`,
		},
		{
			name:   "several roots",
			layout: []string{"testdata/vm.txtar", "r/demo/foo/foo.txtar", "r/demo/foo/skipped.gno"},
			roots:  []string{"testdata", "r"},
			want:   []string{"testdata/vm.txtar", "r/demo/foo/foo.txtar"}, // root order
		},
		{
			// The roots share one subtest namespace, so the check spans them.
			name:        "duplicate base names across roots",
			layout:      []string{"testdata/wugnot.txtar", "r/demo/wugnot/wugnot.txtar"},
			roots:       []string{"testdata", "r"},
			errContains: `duplicate testscript base name "wugnot.txtar"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			for _, rel := range tt.layout {
				fpath := filepath.Join(dir, filepath.FromSlash(rel))
				require.NoError(t, os.MkdirAll(filepath.Dir(fpath), 0o755))
				require.NoError(t, os.WriteFile(fpath, nil, 0o644))
			}
			abs := func(rels []string) []string {
				var out []string
				for _, rel := range rels {
					out = append(out, filepath.Join(dir, filepath.FromSlash(rel)))
				}
				return out
			}

			roots := abs(tt.roots)
			if roots == nil {
				roots = []string{dir}
			}

			got, err := FindTestScripts(roots...)
			if tt.errContains != "" {
				assert.ErrorContains(t, err, tt.errContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, abs(tt.want), got)
		})
	}
}
