package gnomod

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleDeprecated(t *testing.T) {
	for _, tc := range []struct {
		desc, in, expected string
	}{
		{
			desc: "no_comment",
			in:   `module m`,
		},
		{
			desc: "other_comment",
			in: `// yo
			module m`,
		},
		{
			desc: "deprecated_no_colon",
			in: `//Deprecated
			module m`,
		},
		{
			desc: "deprecated_no_space",
			in: `//Deprecated:blah
			module m`,
			expected: "blah",
		},
		{
			desc: "deprecated_simple",
			in: `// Deprecated: blah
			module m`,
			expected: "blah",
		},
		{
			desc: "deprecated_lowercase",
			in: `// deprecated: blah
			module m`,
		},
		{
			desc: "deprecated_multiline",
			in: `// Deprecated: one
			// two
			module m`,
			expected: "one\ntwo",
		},
		{
			desc: "deprecated_mixed",
			in: `// some other comment
			// Deprecated: blah
			module m`,
		},
		{
			desc: "deprecated_middle",
			in: `// module m is Deprecated: blah
			module m`,
		},
		{
			desc: "deprecated_multiple",
			in: `// Deprecated: a
			// Deprecated: b
			module m`,
			expected: "a\nDeprecated: b",
		},
		{
			desc: "deprecated_paragraph",
			in: `// Deprecated: a
			// b
			//
			// c
			module m`,
			expected: "a\nb",
		},
		{
			desc: "deprecated_paragraph_space",
			in: `// Deprecated: the next line has a space
			//
			// c
			module m`,
			expected: "the next line has a space",
		},
		{
			desc:     "deprecated_suffix",
			in:       `module m // Deprecated: blah`,
			expected: "blah",
		},
		{
			desc: `deprecated_mixed_suffix`,
			in: `// some other comment
			module m // Deprecated: blah`,
		},
		{
			desc: "deprecated_mixed_suffix_paragraph",
			in: `// some other comment
			//
			module m // Deprecated: blah`,
			expected: "blah",
		},
		{
			desc: "deprecated_block",
			in: `// Deprecated: blah
			module (
				m
			)`,
			expected: "blah",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			f, err := ParseBytes("in", []byte(tc.in))
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, f.Module.Deprecated)
		})
	}
}

func TestParseDraft(t *testing.T) {
	for _, tc := range []struct {
		desc, in string
		expected bool
	}{
		{
			desc: "no_comment",
			in:   `module m`,
		},
		{
			desc: "other_comment",
			in:   `// yo`,
		},
		{
			desc:     "draft_no_space",
			in:       `//Draft`,
			expected: true,
		},
		{
			desc:     "draft_simple",
			in:       `// Draft`,
			expected: true,
		},
		{
			desc: "draft_lowercase",
			in:   `// draft`,
		},
		{
			desc: "draft_multiline",
			in: `// Draft
			// yo`,
		},
		{
			desc: "draft_mixed",
			in: `// some other comment
			// Draft`,
		},
		{
			desc: "draft_not_first_line",
			in: `
			// Draft`,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			f, err := ParseBytes("in", []byte(tc.in))
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, f.Draft)
		})
	}
}

func TestParseFilepath(t *testing.T) {
	pkgDir := "bar"
	for _, tc := range []struct {
		desc, modData, modPath, errShouldContain string
	}{
		{
			desc:             "file not exists",
			modData:          `module foo`,
			modPath:          filepath.Join(pkgDir, "mod.gno"),
			errShouldContain: "could not read gno.mod file:",
		},
		{
			desc:             "file path is dir",
			modData:          `module foo`,
			modPath:          pkgDir,
			errShouldContain: "is a directory",
		},
		{
			desc:    "valid gno.mod file",
			modData: `module foo`,
			modPath: filepath.Join(pkgDir, "gno.mod"),
		},
		{
			desc: "valid gno.mod file with replace",
			modData: `module foo
			replace bar => ../bar`,
			modPath: filepath.Join(pkgDir, "gno.mod"),
		},
		{
			desc:             "error bad module directive",
			modData:          `module foo v0.0.0`,
			modPath:          filepath.Join(pkgDir, "gno.mod"),
			errShouldContain: "error parsing gno.mod file at",
		},
		{
			desc:             "error gno.mod without module",
			modData:          `replace bar => ../bar`,
			modPath:          filepath.Join(pkgDir, "gno.mod"),
			errShouldContain: "requires module",
		},
		{
			desc: "error gno.mod with require",
			modData: `module foo
			require bar v0.0.0`,
			modPath:          filepath.Join(pkgDir, "gno.mod"),
			errShouldContain: "unknown directive: require",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			// Create test dir
			tempDir, cleanUpFn := testutils.NewTestCaseDir(t)
			require.NotNil(t, tempDir)
			defer cleanUpFn()

			// Create gno package
			createGnoModPkg(t, tempDir, pkgDir, tc.modData)

			_, err := ParseFilepath(filepath.Join(tempDir, tc.modPath))
			if tc.errShouldContain != "" {
				assert.ErrorContains(t, err, tc.errShouldContain)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestParseWithInvalidModulePath tests that Parse correctly rejects gno.mod files
// with invalid module paths.
func TestParseWithInvalidModulePath(t *testing.T) {
	tests := []struct {
		name    string
		modData string
		errMsg  string
	}{
		{
			name:    "valid module path",
			modData: "module gno.land/p/demo/foo",
			errMsg:  "",
		},
		{
			name:    "module path with space",
			modData: "module \"gno.land/p/demo/ foo\"",
			errMsg:  "malformed import path \"gno.land/p/demo/ foo\": invalid char ' '",
		},
		{
			name:    "module path with Unicode",
			modData: "module gno.land/p/demo/한글",
			errMsg:  "malformed import path \"gno.land/p/demo/한글\": invalid char '한'",
		},
		{
			name:    "module path with invalid character",
			modData: "module gno.land/p/demo/foo*bar",
			errMsg:  "malformed import path \"gno.land/p/demo/foo*bar\": invalid char '*'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseBytes("gno.mod", []byte(tt.modData))
			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCreateGnoModFileWithInvalidPath tests that CreateGnoModFile correctly rejects
// invalid module paths.
func TestCreateGnoModFileWithInvalidPath(t *testing.T) {
	tests := []struct {
		name    string
		modPath string
		errMsg  string
	}{
		{
			name:    "valid module path",
			modPath: "gno.land/p/demo/foo",
			errMsg:  "",
		},
		{
			name:    "module path with space",
			modPath: "gno.land/p/demo/ foo",
			errMsg:  "malformed import path \"gno.land/p/demo/ foo\": invalid char ' '",
		},
		{
			name:    "module path with Unicode",
			modPath: "gno.land/p/demo/한글",
			errMsg:  "malformed import path \"gno.land/p/demo/한글\": invalid char '한'",
		},
		{
			name:    "module path with invalid character",
			modPath: "gno.land/p/demo/foo*bar",
			errMsg:  "malformed import path \"gno.land/p/demo/foo*bar\": invalid char '*'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanUpFn := testutils.NewTestCaseDir(t)
			defer cleanUpFn()

			err := CreateGnoModFile(tempDir, tt.modPath)
			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				if err != nil && !assert.Contains(t, err.Error(), "dir") {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
