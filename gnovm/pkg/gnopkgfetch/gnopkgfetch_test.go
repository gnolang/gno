package gnopkgfetch

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func TestFetchDeps(t *testing.T) {
	for _, tc := range []struct {
		desc               string
		pkgPath            string
		modFile            gnomod.File
		errorShouldContain string
		requirements       []string
		ioErrContains      []string
	}{
		{
			desc:    "not_exists",
			pkgPath: "gno.land/p/demo/does_not_exists",
			modFile: gnomod.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
			},
			errorShouldContain: "failed to query files list for pkg \"gno.land/p/demo/does_not_exists\": package \"gno.land/p/demo/does_not_exists\" is not available",
		}, {
			desc:    "fetch_gno.land/p/demo/avl",
			pkgPath: "gno.land/p/demo/avl",
			modFile: gnomod.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
			},
			requirements: []string{"avl"},
			ioErrContains: []string{
				"gno: downloading gno.land/p/demo/avl",
			},
		}, {
			desc:    "fetch_gno.land/p/demo/blog6",
			pkgPath: "gno.land/p/demo/blog",
			modFile: gnomod.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
			},
			requirements: []string{"avl", "blog", "ufmt", "mux"},
			ioErrContains: []string{
				"gno: downloading gno.land/p/demo/blog",
				"gno: downloading gno.land/p/demo/avl",
				"gno: downloading gno.land/p/demo/ufmt",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			mockErr := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetErr(commands.WriteNopCloser(mockErr))

			InjectExamplesClient(t)

			// Create test dir
			dirPath, cleanUpFn := testutils.NewTestCaseDir(t)
			assert.NotNil(t, dirPath)
			defer cleanUpFn()

			err := os.WriteFile(filepath.Join(dirPath, "main.gno"), []byte(fmt.Sprintf("package main\n\n import %q\n", tc.pkgPath)), 0644)
			require.NoError(t, err)

			tmpGnoHome, err := os.MkdirTemp(os.TempDir(), "gnotesthome_")
			require.NoError(t, err)
			t.Cleanup(func() { os.RemoveAll(tmpGnoHome) })
			t.Setenv("GNOHOME", tmpGnoHome)

			// gno: downloading dependencies
			err = FetchPackageImportsRecursively(io, dirPath, &tc.modFile)
			if tc.errorShouldContain != "" {
				require.ErrorContains(t, err, tc.errorShouldContain)
			} else {
				require.Nil(t, err)

				// Read dir
				entries, err := os.ReadDir(filepath.Join(tmpGnoHome, "pkg", "mod", "gno.land", "p", "demo"))
				require.Nil(t, err)

				// Check dir entries
				assert.Equal(t, len(tc.requirements), len(entries))
				for _, e := range entries {
					assert.Contains(t, tc.requirements, e.Name())
				}

				// Check logs
				for _, c := range tc.ioErrContains {
					assert.Contains(t, mockErr.String(), c)
				}

				mockErr.Reset()

				// Try gno: downloading again. Should be cached
				FetchPackageImportsRecursively(io, dirPath, &tc.modFile)
				for _, c := range tc.ioErrContains {
					assert.NotContains(t, mockErr.String(), c)
				}
			}
		})
	}
}
