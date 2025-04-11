package gnomod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGnoModFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		desc             string
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
			errShouldContain: "malformed import path",
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
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			// Create test dir
			dirPath := t.TempDir()

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
