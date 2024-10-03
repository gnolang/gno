package gnomod

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

const testRemote string = "test3.gno.land:26657"

func TestFetchDeps(t *testing.T) {
	for _, tc := range []struct {
		desc                 string
		modFile              File
		errorShouldContain   string
		requirements         []string
		stdOutContains       []string
		cachedStdOutContains []string
	}{
		{
			desc: "not_exists",
			modFile: File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
				Require: []*modfile.Require{
					{
						Mod: module.Version{
							Path:    "gno.land/p/demo/does_not_exists",
							Version: "v0.0.0",
						},
					},
				},
			},
			errorShouldContain: "querychain (gno.land/p/demo/does_not_exists)",
		}, {
			desc: "fetch_gno.land/p/demo/avl",
			modFile: File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
				Require: []*modfile.Require{
					{
						Mod: module.Version{
							Path:    "gno.land/p/demo/avl",
							Version: "v0.0.0",
						},
					},
				},
			},
			requirements: []string{"avl"},
			stdOutContains: []string{
				"fetching gno.land/p/demo/avl",
			},
			cachedStdOutContains: []string{
				"cached gno.land/p/demo/avl",
			},
		}, {
			desc: "fetch_gno.land/p/demo/blog",
			modFile: File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path: "testFetchDeps",
					},
				},
				Require: []*modfile.Require{
					{
						Mod: module.Version{
							Path:    "gno.land/p/demo/blog",
							Version: "v0.0.0",
						},
					},
				},
			},
			requirements: []string{"avl", "blog", "ufmt"},
			stdOutContains: []string{
				"fetching gno.land/p/demo/blog",
				"fetching gno.land/p/demo/avl // indirect",
				"fetching gno.land/p/demo/ufmt // indirect",
			},
			cachedStdOutContains: []string{
				"cached gno.land/p/demo/blog",
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer func() {
				log.SetOutput(os.Stderr)
			}()

			// Create test dir
			dirPath, cleanUpFn := testutils.NewTestCaseDir(t)
			assert.NotNil(t, dirPath)
			defer cleanUpFn()

			// Fetching dependencies
			err := tc.modFile.FetchDeps(dirPath, testRemote, true)
			if tc.errorShouldContain != "" {
				require.ErrorContains(t, err, tc.errorShouldContain)
			} else {
				require.Nil(t, err)

				// Read dir
				entries, err := os.ReadDir(filepath.Join(dirPath, "gno.land", "p", "demo"))
				require.Nil(t, err)

				// Check dir entries
				assert.Equal(t, len(tc.requirements), len(entries))
				for _, e := range entries {
					assert.Contains(t, tc.requirements, e.Name())
				}

				// Check logs
				for _, c := range tc.stdOutContains {
					assert.Contains(t, buf.String(), c)
				}

				buf.Reset()

				// Try fetching again. Should be cached
				tc.modFile.FetchDeps(dirPath, testRemote, true)
				for _, c := range tc.cachedStdOutContains {
					assert.Contains(t, buf.String(), c)
				}
			}
		})
	}
}
