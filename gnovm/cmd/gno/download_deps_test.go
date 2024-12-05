package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/pkgdownload/examplespkgfetcher"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func TestDownloadDeps(t *testing.T) {
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
			errorShouldContain: "query files list for pkg \"gno.land/p/demo/does_not_exists\": package \"gno.land/p/demo/does_not_exists\" is not available",
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
		}, {
			desc:    "fetch_replace_gno.land/p/demo/avl",
			pkgPath: "gno.land/p/demo/replaced_avl",
			modFile: gnomod.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
				Replace: []*modfile.Replace{{
					Old: module.Version{Path: "gno.land/p/demo/replaced_avl"},
					New: module.Version{Path: "gno.land/p/demo/avl"},
				}},
			},
			requirements: []string{"avl"},
			ioErrContains: []string{
				"gno: downloading gno.land/p/demo/avl",
			},
		}, {
			desc:    "fetch_replace_local",
			pkgPath: "gno.land/p/demo/foo",
			modFile: gnomod.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
				Replace: []*modfile.Replace{{
					Old: module.Version{Path: "gno.land/p/demo/foo"},
					New: module.Version{Path: "../local_foo"},
				}},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			mockErr := bytes.NewBufferString("")
			io := commands.NewTestIO()
			io.SetErr(commands.WriteNopCloser(mockErr))

			dirPath := t.TempDir()

			err := os.WriteFile(filepath.Join(dirPath, "main.gno"), []byte(fmt.Sprintf("package main\n\n import %q\n", tc.pkgPath)), 0o644)
			require.NoError(t, err)

			tmpGnoHome := t.TempDir()
			t.Setenv("GNOHOME", tmpGnoHome)

			fetcher := examplespkgfetcher.New()

			// gno: downloading dependencies
			err = downloadDeps(io, dirPath, &tc.modFile, fetcher)
			if tc.errorShouldContain != "" {
				require.ErrorContains(t, err, tc.errorShouldContain)
			} else {
				require.Nil(t, err)

				// Read dir
				entries, err := os.ReadDir(filepath.Join(tmpGnoHome, "pkg", "mod", "gno.land", "p", "demo"))
				if !os.IsNotExist(err) {
					require.Nil(t, err)
				}

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

				// Try fetching again. Should be cached
				downloadDeps(io, dirPath, &tc.modFile, fetcher)
				for _, c := range tc.ioErrContains {
					assert.NotContains(t, mockErr.String(), c)
				}
			}
		})
	}
}
