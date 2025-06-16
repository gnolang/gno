package gnomod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/testutils"
)

func TestFile_GetGno(t *testing.T) {
	testCases := []struct {
		name     string
		file     *File
		expected string
	}{
		{
			name:     "empty version returns default",
			file:     &File{},
			expected: "0.0",
		},
		{
			name: "custom version",
			file: func() *File {
				f := &File{}
				f.Gno = "0.9"
				return f
			}(),
			expected: "0.9",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version := tc.file.GetGno()
			assert.Equal(t, tc.expected, version)
		})
	}
}

func TestFile_SetGno(t *testing.T) {
	file := &File{}
	file.SetGno("0.9")
	assert.Equal(t, "0.9", file.Gno)
}

func TestFile_AddReplace(t *testing.T) {
	testCases := []struct {
		name           string
		initialFile    *File
		oldPath        string
		newPath        string
		expectedCount  int
		expectedFirst  Replace
		expectedSecond Replace
	}{
		{
			name:          "add new replace",
			initialFile:   &File{},
			oldPath:       "gno.land/p/demo/foo",
			newPath:       "gno.land/p/demo/bar",
			expectedCount: 1,
			expectedFirst: Replace{
				Old: "gno.land/p/demo/foo",
				New: "gno.land/p/demo/bar",
			},
		},
		{
			name: "update existing replace",
			initialFile: func() *File {
				f := &File{}
				f.Replace = []Replace{
					{
						Old: "gno.land/p/demo/foo",
						New: "gno.land/p/demo/old",
					},
				}
				return f
			}(),
			oldPath:       "gno.land/p/demo/foo",
			newPath:       "gno.land/p/demo/new",
			expectedCount: 1,
			expectedFirst: Replace{
				Old: "gno.land/p/demo/foo",
				New: "gno.land/p/demo/new",
			},
		},
		{
			name: "add second replace",
			initialFile: func() *File {
				f := &File{}
				f.Replace = []Replace{
					{
						Old: "gno.land/p/demo/foo",
						New: "gno.land/p/demo/bar",
					},
				}
				return f
			}(),
			oldPath:       "gno.land/p/demo/baz",
			newPath:       "gno.land/p/demo/qux",
			expectedCount: 2,
			expectedFirst: Replace{
				Old: "gno.land/p/demo/foo",
				New: "gno.land/p/demo/bar",
			},
			expectedSecond: Replace{
				Old: "gno.land/p/demo/baz",
				New: "gno.land/p/demo/qux",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.initialFile.AddReplace(tc.oldPath, tc.newPath)
			assert.Equal(t, tc.expectedCount, len(tc.initialFile.Replace))
			if tc.expectedCount > 0 {
				assert.Equal(t, tc.expectedFirst, tc.initialFile.Replace[0])
			}
			if tc.expectedCount > 1 {
				assert.Equal(t, tc.expectedSecond, tc.initialFile.Replace[1])
			}
		})
	}
}

func TestFile_DropReplace(t *testing.T) {
	testCases := []struct {
		name          string
		initialFile   *File
		oldPath       string
		expectedCount int
		expectedFirst Replace
	}{
		{
			name: "drop existing replace",
			initialFile: func() *File {
				f := &File{}
				f.Replace = []Replace{
					{
						Old: "gno.land/p/demo/foo",
						New: "gno.land/p/demo/bar",
					},
					{
						Old: "gno.land/p/demo/baz",
						New: "gno.land/p/demo/qux",
					},
				}
				return f
			}(),
			oldPath:       "gno.land/p/demo/foo",
			expectedCount: 1,
			expectedFirst: Replace{
				Old: "gno.land/p/demo/baz",
				New: "gno.land/p/demo/qux",
			},
		},
		{
			name: "drop non-existent replace",
			initialFile: func() *File {
				f := &File{}
				f.Replace = []Replace{
					{
						Old: "gno.land/p/demo/foo",
						New: "gno.land/p/demo/bar",
					},
				}
				return f
			}(),
			oldPath:       "gno.land/p/demo/baz",
			expectedCount: 1,
			expectedFirst: Replace{
				Old: "gno.land/p/demo/foo",
				New: "gno.land/p/demo/bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.initialFile.DropReplace(tc.oldPath)
			assert.Equal(t, tc.expectedCount, len(tc.initialFile.Replace))
			if tc.expectedCount > 0 {
				assert.Equal(t, tc.expectedFirst, tc.initialFile.Replace[0])
			}
		})
	}
}

func TestFile_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		file        *File
		expectedErr string
	}{
		{
			name: "valid module path",
			file: func() *File {
				f := &File{}
				f.Module = "gno.land/p/demo/foo"
				return f
			}(),
		},
		{
			name:        "empty module path",
			file:        &File{},
			expectedErr: "invalid gnomod.toml: 'module' is required",
		},
		{
			name: "invalid module path with space",
			file: func() *File {
				f := &File{}
				f.Module = "gno.land/p/demo/ foo"
				return f
			}(),
			expectedErr: "malformed import path",
		},
		{
			name: "invalid module path with Unicode",
			file: func() *File {
				f := &File{}
				f.Module = "gno.land/p/demo/한글"
				return f
			}(),
			expectedErr: "malformed import path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.file.Validate()
			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestFile_Resolve(t *testing.T) {
	testCases := []struct {
		name     string
		file     *File
		target   string
		expected string
	}{
		{
			name: "resolve with replace",
			file: func() *File {
				f := &File{}
				f.Replace = []Replace{
					{
						Old: "gno.land/p/demo/foo",
						New: "gno.land/p/demo/bar",
					},
				}
				return f
			}(),
			target:   "gno.land/p/demo/foo",
			expected: "gno.land/p/demo/bar",
		},
		{
			name:     "resolve without replace",
			file:     &File{},
			target:   "gno.land/p/demo/foo",
			expected: "gno.land/p/demo/foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolved := tc.file.Resolve(tc.target)
			assert.Equal(t, tc.expected, resolved)
		})
	}
}

func TestFile_WriteFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, cleanUpFn := testutils.NewTestCaseDir(t)
	require.NotNil(t, tempDir)
	defer cleanUpFn()

	testCases := []struct {
		name        string
		file        *File
		expectedErr string
	}{
		{
			name: "write valid file",
			file: func() *File {
				f := &File{}
				f.Module = "gno.land/p/demo/foo"
				return f
			}(),
		},
		{
			name: "write to non-existent directory",
			file: func() *File {
				f := &File{}
				f.Module = "gno.land/p/demo/foo"
				return f
			}(),
			expectedErr: "no such file or directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var fpath string
			if tc.expectedErr == "" {
				fpath = filepath.Join(tempDir, "gnomod.toml")
			} else {
				fpath = filepath.Join(tempDir, "nonexistent", "gnomod.toml")
			}

			err := tc.file.WriteFile(fpath)
			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				return
			}

			assert.NoError(t, err)
			_, err = os.Stat(fpath)
			assert.NoError(t, err)
		})
	}
}

func TestFile_Sanitize(t *testing.T) {
	testCases := []struct {
		name     string
		file     *File
		expected *File
	}{
		{
			name: "sanitize empty version",
			file: &File{},
			expected: func() *File {
				f := &File{}
				f.Gno = "0.0"
				f.Replace = []Replace{}
				return f
			}(),
		},
		{
			name: "sanitize empty replaces",
			file: func() *File {
				f := &File{}
				f.Replace = []Replace{
					{Old: "", New: "bar"},
					{Old: "foo", New: ""},
					{Old: "baz", New: "baz"},
				}
				return f
			}(),
			expected: func() *File {
				f := &File{}
				f.Gno = "0.0"
				f.Replace = []Replace{}
				return f
			}(),
		},
		{
			name: "sanitize duplicate replaces",
			file: func() *File {
				f := &File{}
				f.Replace = []Replace{
					{Old: "foo", New: "bar"},
					{Old: "foo", New: "baz"},
				}
				return f
			}(),
			expected: func() *File {
				f := &File{}
				f.Gno = "0.0"
				f.Replace = []Replace{
					{Old: "foo", New: "bar"},
				}
				return f
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.file.Sanitize()
			assert.Equal(t, tc.expected.Gno, tc.file.Gno)
			assert.Equal(t, tc.expected.Replace, tc.file.Replace)
		})
	}
}
