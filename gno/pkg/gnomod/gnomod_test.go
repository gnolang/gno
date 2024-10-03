package gnomod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGnoModFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		desc             string
		in               []struct{ filename, content string }
		inModPath        string
		out              string
		errShouldContain string
	}{
		{
			desc:      "empty directory",
			inModPath: "gno.land/p/demo/foo",
			out:       "module gno.land/p/demo/foo\n",
		},
		{
			desc:             "empty directory (without modPath)",
			errShouldContain: "cannot determine package name",
		},
		{
			desc:             "invalid modPath 1",
			inModPath:        " ",
			errShouldContain: "malformed import path",
		},
		{
			desc:             "invalid modPath 2",
			inModPath:        "\"",
			errShouldContain: "malformed import path",
		},
		{
			desc: "valid package",
			in: []struct{ filename, content string }{
				{
					"foo.gno",
					`package foo`,
				},
			},
			inModPath: "gno.land/p/demo/foo",
			out:       "module gno.land/p/demo/foo\n",
		},
		{
			desc: "valid package (without modPath)",
			in: []struct{ filename, content string }{
				{
					"foo.gno",
					`package foo`,
				},
			},
			out: "module foo\n",
		},
		{
			desc: "ambigious package names",
			in: []struct{ filename, content string }{
				{
					"foo.gno",
					`package foo`,
				},
				{
					"bar.gno",
					`package bar`,
				},
			},
			inModPath: "gno.land/p/demo/foo",
			out:       "module gno.land/p/demo/foo\n",
		},
		{
			desc: "ambigious package names (without modPath)",
			in: []struct{ filename, content string }{
				{
					"foo.gno",
					`package foo`,
				},
				{
					"bar.gno",
					`package bar`,
				},
			},
			errShouldContain: "package name mismatch:",
		},
		{
			desc: "valid package with gno.mod file",
			in: []struct{ filename, content string }{
				{
					"foo.gno",
					`package foo`,
				},
				{
					"gno.mod",
					`module gno.land/p/demo/foo`,
				},
			},
			inModPath:        "gno.land/p/demo/foo",
			errShouldContain: "gno.mod file already exists",
		},
		{
			desc: "valid package with gno.mod file (without modPath)",
			in: []struct{ filename, content string }{
				{
					"foo.gno",
					`package foo`,
				},
				{
					"gno.mod",
					`module gno.land/p/demo/foo`,
				},
			},
			errShouldContain: "gno.mod file already exists",
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			// Create test dir
			dirPath, cleanUpFn := testutils.NewTestCaseDir(t)
			require.NotNil(t, dirPath)
			defer cleanUpFn()

			// Create files
			for _, f := range tc.in {
				err := os.WriteFile(filepath.Join(dirPath, f.filename), []byte(f.content), 0o644)
				require.NoError(t, err)
			}

			err := CreateGnoModFile(dirPath, tc.inModPath)
			if tc.errShouldContain != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errShouldContain)
				return
			}
			assert.NoError(t, err)

			// Verify gno.mod file
			bz, err := os.ReadFile(filepath.Join(dirPath, "gno.mod"))
			assert.NoError(t, err)
			assert.Equal(t, tc.out, string(bz))
		})
	}
}
