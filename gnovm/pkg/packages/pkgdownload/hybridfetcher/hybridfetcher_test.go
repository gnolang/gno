package hybridfetcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHybridFetcher_FetchPackage(t *testing.T) {
	tempDir := t.TempDir()

	testPkgDir := filepath.Join(tempDir, "p", "test", "pkg")
	require.NoError(t, os.MkdirAll(testPkgDir, 0755))

	// Create a simple test file
	testFile := filepath.Join(testPkgDir, "test.gno")
	testContent := []byte(`package pkg

func Hello() string {
	return "Hello from test"
}`)
	require.NoError(t, os.WriteFile(testFile, testContent, 0644))

	gnoModFile := filepath.Join(testPkgDir, "gnomod.toml")
	gnoModContent := []byte(`module = "gno.land/r/examples_ignored"`)
	require.NoError(t, os.WriteFile(gnoModFile, gnoModContent, 0644))

	tests := []struct {
		name          string
		pkgPath       string
		setupFetcher  func() *HybridFetcher
		expectSuccess bool
		expectFiles   int
	}{
		{
			name:    "fetch local package successfully",
			pkgPath: "gno.land/p/test/pkg",
			setupFetcher: func() *HybridFetcher {
				hf := &HybridFetcher{
					localFetchers: []pkgdownload.PackageFetcher{
						&localDirFetcher{baseDir: tempDir},
					},
					rpcFetcher: nil, // Don't need RPC for this test
					verbose:    true,
				}
				return hf
			},
			expectSuccess: true,
			expectFiles:   2, // test.gno and gnomod
		},
		{
			name:    "fallback to RPC when not found locally",
			pkgPath: "gno.land/p/notexist/pkg",
			setupFetcher: func() *HybridFetcher {
				// Mock RPC fetcher that returns an error
				mockRPC := &mockPackageFetcher{
					fetchFunc: func(pkgPath string) ([]*std.MemFile, error) {
						return nil, assert.AnError
					},
				}

				hf := &HybridFetcher{
					localFetchers: []pkgdownload.PackageFetcher{
						&localDirFetcher{baseDir: tempDir},
					},
					rpcFetcher: mockRPC,
					verbose:    false,
				}
				return hf
			},
			expectSuccess: false,
			expectFiles:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hf := tt.setupFetcher()
			files, err := hf.FetchPackage(tt.pkgPath)

			if tt.expectSuccess {
				require.NoError(t, err)
				assert.Len(t, files, tt.expectFiles)
				if tt.expectFiles > 0 {
					// Find the test.gno file
					var testFile *std.MemFile
					for _, f := range files {
						if f.Name == "test.gno" {
							testFile = f
							break
						}
					}
					require.NotNil(t, testFile, "test.gno file should exist")
					assert.Contains(t, testFile.Body, "Hello from test")
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

type mockPackageFetcher struct {
	fetchFunc func(pkgPath string) ([]*std.MemFile, error)
}

func (m *mockPackageFetcher) FetchPackage(pkgPath string) ([]*std.MemFile, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(pkgPath)
	}
	return nil, nil
}

var _ pkgdownload.PackageFetcher = (*HybridFetcher)(nil)
var _ pkgdownload.PackageFetcher = (*localDirFetcher)(nil)
var _ pkgdownload.PackageFetcher = (*mockPackageFetcher)(nil)
