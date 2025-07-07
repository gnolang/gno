package gnomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestParseBytes tests parsing of both gno.mod and gnomod.toml files
func TestParseBytes(t *testing.T) {
	testCases := []struct {
		name            string
		content         string
		fileType        string // "gno.mod" or "gnomod.toml"
		expectedModule  string
		expectedVersion string
		expectedIgnore  bool
		expectedDraft   bool
		expectedError   string
	}{
		// Valid gno.mod cases
		{
			name:            "valid gno.mod with module",
			content:         "module gno.land/p/demo/foo",
			fileType:        "gno.mod",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name:            "valid gno.mod with module and gno version",
			content:         "module gno.land/p/demo/foo\ngno 0.9",
			fileType:        "gno.mod",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.9",
		},
		{
			name:            "valid gno.mod with module and replace",
			content:         "module gno.land/p/demo/foo\nreplace bar => ../bar",
			fileType:        "gno.mod",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name:           "gno.mod with ignore comment",
			content:        "// Ignore\n\nmodule gno.land/p/demo/foo",
			fileType:       "gno.mod",
			expectedModule: "gno.land/p/demo/foo",
			expectedIgnore: true,
		},
		{
			name:           "gno.mod with deprecated comment",
			content:        "// Deprecated: use new module\nmodule gno.land/p/demo/foo",
			fileType:       "gno.mod",
			expectedModule: "gno.land/p/demo/foo",
		},

		// Valid gnomod.toml cases
		{
			name:            "valid gnomod.toml with module",
			content:         "module = \"gno.land/p/demo/foo\"",
			fileType:        "gnomod.toml",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name:            "valid gnomod.toml with module and gno version",
			content:         "module = \"gno.land/p/demo/foo\"\ngno = \"0.9\"",
			fileType:        "gnomod.toml",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.9",
		},
		{
			name:            "valid gnomod.toml with module and replace",
			content:         "module = \"gno.land/p/demo/foo\"\nignore = true",
			fileType:        "gnomod.toml",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
			expectedIgnore:  true,
		},
		{
			name:           "gnomod.toml with ignore flag",
			content:        "module = \"gno.land/p/demo/foo\"\nignore = true",
			fileType:       "gnomod.toml",
			expectedModule: "gno.land/p/demo/foo",
			expectedIgnore: true,
		},
		{
			name:           "gnomod.toml with draft and ignore flags",
			content:        "module = \"gno.land/p/demo/foo\"\ndraft = true\nignore = true",
			fileType:       "gnomod.toml",
			expectedModule: "gno.land/p/demo/foo",
			expectedDraft:  true,
			expectedIgnore: true,
		},

		// Invalid cases
		{
			name:          "invalid gno.mod without module",
			content:       "replace bar => ../bar",
			fileType:      "gno.mod",
			expectedError: "invalid gnomod.toml: 'module' is required",
		},
		{
			name:          "invalid gno.mod with require",
			content:       "module foo\nrequire bar v0.0.0",
			fileType:      "gno.mod",
			expectedError: "unknown directive: require",
		},
		{
			name:          "invalid gnomod.toml without module",
			content:       "gno = \"0.9\"",
			fileType:      "gnomod.toml",
			expectedError: "invalid gnomod.toml: 'module' is required",
		},
		{
			name:          "invalid gnomod.toml with invalid toml",
			content:       "path = gno.land/p/demo/foo",
			fileType:      "gnomod.toml",
			expectedError: "error parsing gnomod.toml file",
		},
		{
			name:          "invalid module path with space",
			content:       "module \"gno.land/p/demo/ foo\"",
			fileType:      "gno.mod",
			expectedError: "malformed import path",
		},
		{
			name:          "invalid module path with Unicode",
			content:       "module gno.land/p/demo/한글",
			fileType:      "gno.mod",
			expectedError: "malformed import path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var file *File
			var err error

			file, err = ParseBytes(tc.fileType, []byte(tc.content))
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedModule, file.Module)
			if tc.expectedVersion != "" {
				assert.Equal(t, tc.expectedVersion, file.GetGno())
			}
			assert.Equal(t, tc.expectedIgnore, file.Ignore)
			assert.Equal(t, tc.expectedDraft, file.Draft)
		})
	}
}

// TestParseMemPackage tests parsing of module files from MemPackage
func TestParseMemPackage(t *testing.T) {
	t.Skip("skipping")
	testCases := []struct {
		name            string
		files           []*std.MemFile
		expectedModule  string
		expectedVersion string
		expectedError   string
	}{
		{
			name: "valid gno.mod in mem package",
			files: []*std.MemFile{
				{Name: "gno.mod", Body: "module gno.land/p/demo/foo"},
			},
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name: "valid gnomod.toml in mem package",
			files: []*std.MemFile{
				{Name: "gnomod.toml", Body: "module = \"gno.land/p/demo/foo\""},
			},
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name: "both files present, prefers gnomod.toml",
			files: []*std.MemFile{
				{Name: "gno.mod", Body: "module gno.land/p/demo/old"},
				{Name: "gnomod.toml", Body: "module = \"gno.land/p/demo/new\""},
			},
			expectedModule:  "gno.land/p/demo/new",
			expectedVersion: "0.0",
		},
		{
			name:          "no module files",
			files:         []*std.MemFile{},
			expectedError: "gnomod.toml not in mem package",
		},
		{
			name: "invalid gno.mod",
			files: []*std.MemFile{
				{Name: "gno.mod", Body: "invalid content"},
			},
			expectedError: "error parsing gno.mod file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mpkg := &std.MemPackage{
				Name:  "test",
				Path:  "gno.land/p/demo/test",
				Files: tc.files,
			}

			file, err := ParseMemPackage(mpkg)
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedModule, file.Module)
			assert.Equal(t, tc.expectedVersion, file.GetGno())
		})
	}
}
