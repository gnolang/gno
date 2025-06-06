package gnomod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRootDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, cleanUpFn := testutils.NewTestCaseDir(t)
	require.NotNil(t, tempDir)
	defer cleanUpFn()

	// Create a directory structure for testing
	// tempDir/
	// ├── subdir1/
	// │   ├── gnomod.toml
	// │   └── subdir2/
	// │       └── subdir3/
	// │           └── gno.mod
	// └── subdir4/
	//     └── subdir5/
	//         └── no_mod_here

	// Create directories
	subdir1 := filepath.Join(tempDir, "subdir1")
	subdir2 := filepath.Join(subdir1, "subdir2")
	subdir3 := filepath.Join(subdir2, "subdir3")
	subdir4 := filepath.Join(tempDir, "subdir4")
	subdir5 := filepath.Join(subdir4, "subdir5")

	err := os.MkdirAll(subdir3, 0o755)
	require.NoError(t, err)
	err = os.MkdirAll(subdir5, 0o755)
	require.NoError(t, err)

	// Create gnomod.toml in subdir1
	err = os.WriteFile(filepath.Join(subdir1, "gnomod.toml"), []byte("[module]\npath = \"gno.land/p/demo/root\""), 0o644)
	require.NoError(t, err)

	// Create gno.mod in subdir3
	err = os.WriteFile(filepath.Join(subdir3, "gno.mod"), []byte("module gno.land/p/demo/subdir"), 0o644)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		path        string
		expected    string
		expectedErr string
	}{
		{
			name:     "find root with gnomod.toml",
			path:     subdir1,
			expected: subdir1,
		},
		{
			name:     "find root from subdir with gnomod.toml",
			path:     subdir2,
			expected: subdir1,
		},
		{
			name:     "find root with gno.mod",
			path:     subdir3,
			expected: subdir3,
		},
		{
			name:        "no mod file found",
			path:        subdir5,
			expectedErr: ErrModFileNotFound.Error(),
		},
		{
			name:        "non-absolute path",
			path:        "relative/path",
			expectedErr: "requires absolute path",
		},
		{
			name:        "empty path",
			path:        "",
			expectedErr: "requires absolute path",
		},
		{
			name:        "root directory",
			path:        "/",
			expectedErr: ErrModFileNotFound.Error(),
		},
		{
			name:        "current directory with dot",
			path:        ".",
			expectedErr: "requires absolute path",
		},
		{
			name:        "parent directory with dot dot",
			path:        "..",
			expectedErr: "requires absolute path",
		},
		{
			name:     "path with trailing slash",
			path:     subdir1 + "/",
			expected: subdir1,
		},
		{
			name:     "path with multiple slashes",
			path:     subdir1 + "///",
			expected: subdir1,
		},
		{
			name:     "path with dot components",
			path:     filepath.Join(subdir1, ".", "subdir2", "..", "subdir2"),
			expected: subdir1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root, err := FindRootDir(tc.path)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, root)
		})
	}

	// Test file system error cases
	t.Run("file system errors", func(t *testing.T) {
		t.Run("directory with no read permissions", func(t *testing.T) {
			// Create a directory with no read permissions
			noReadDir := filepath.Join(tempDir, "no_read")
			err := os.Mkdir(noReadDir, 0o000)
			require.NoError(t, err)
			defer os.Chmod(noReadDir, 0o755) // Restore permissions for cleanup

			// Try to find root from a subdirectory of no_read
			_, err = FindRootDir(filepath.Join(noReadDir, "subdir"))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "permission denied")
		})

		t.Run("directory with no execute permissions", func(t *testing.T) {
			// Create a directory with no execute permissions
			noExecDir := filepath.Join(tempDir, "no_exec")
			err := os.Mkdir(noExecDir, 0o444)
			require.NoError(t, err)
			defer os.Chmod(noExecDir, 0o755) // Restore permissions for cleanup

			// Try to find root from a subdirectory of no_exec
			_, err = FindRootDir(filepath.Join(noExecDir, "subdir"))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "permission denied")
		})

		t.Run("symlink loop", func(t *testing.T) {
			// Create a directory with a symlink loop
			symlinkDir := filepath.Join(tempDir, "symlink")
			err := os.Mkdir(symlinkDir, 0o755)
			require.NoError(t, err)

			// Create a symlink that points to itself
			err = os.Symlink(symlinkDir, filepath.Join(symlinkDir, "loop"))
			require.NoError(t, err)

			// Try to find root from the symlink directory
			_, err = FindRootDir(symlinkDir)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), ErrModFileNotFound.Error())
		})
	})
}
