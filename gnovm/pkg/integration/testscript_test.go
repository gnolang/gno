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
			// testscript's directory mode also globs *.txt; discovery does not,
			// so a bare .txt sitting next to real code stays a fixture.
			name:   "only .txtar is matched",
			layout: []string{"r/demo/foo/foo.txt", "r/demo/foo/foo.txtar"},
			want:   []string{"r/demo/foo/foo.txtar"},
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

func TestNewTestingParamsFromRoots(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	write := func(rel string) string {
		fpath := filepath.Join(dir, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(fpath), 0o755))
		require.NoError(t, os.WriteFile(fpath, nil, 0o644))
		return fpath
	}
	platform := write("testdata/vm.txtar")
	local := write("r/demo/foo/foo.txtar")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "empty"), 0o755))

	t.Run("files list drives the run", func(t *testing.T) {
		t.Parallel()

		p, err := NewTestingParamsFromRoots(t,
			filepath.Join(dir, "testdata"), filepath.Join(dir, "r"))
		require.NoError(t, err)
		// testscript only honors Params.Files when Params.Dir is empty; a Dir
		// default creeping in here would send RunT down its directory branch and
		// leave every discovered root unrun, without an error.
		assert.Empty(t, p.Dir, "a non-empty Dir makes testscript ignore Files")
		assert.Equal(t, []string{platform, local}, p.Files)
	})

	t.Run("a root with no script is an error", func(t *testing.T) {
		t.Parallel()

		// The failure this pins: an aggregate count is satisfied by the other
		// roots, so a root gone empty (a stale GNOROOT, a moved directory) would
		// silently contribute nothing and the suite would still report ok.
		_, err := NewTestingParamsFromRoots(t,
			filepath.Join(dir, "testdata"), filepath.Join(dir, "empty"))
		assert.ErrorContains(t, err, "no testscript file found below")
		assert.ErrorContains(t, err, filepath.Join(dir, "empty"))
	})

	t.Run("no root at all", func(t *testing.T) {
		t.Parallel()

		_, err := NewTestingParamsFromRoots(t)
		assert.ErrorContains(t, err, "no testscript root given")
	})

	t.Run("missing root", func(t *testing.T) {
		t.Parallel()

		_, err := NewTestingParamsFromRoots(t, filepath.Join(dir, "does-not-exist"))
		require.Error(t, err)
	})
}
