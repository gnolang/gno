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
			f, err := Parse("in", []byte(tc.in))
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
			f, err := Parse("in", []byte(tc.in))
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, f.Draft)
		})
	}
}

func TestParseGnoMod(t *testing.T) {
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
			desc:             "error parsing gno.mod",
			modData:          `module foo v0.0.0`,
			modPath:          filepath.Join(pkgDir, "gno.mod"),
			errShouldContain: "error parsing gno.mod file at",
		},
		{
			desc:             "error validating gno.mod",
			modData:          `require bar v0.0.0`,
			modPath:          filepath.Join(pkgDir, "gno.mod"),
			errShouldContain: "error validating gno.mod file at",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			// Create test dir
			tempDir, cleanUpFn := testutils.NewTestCaseDir(t)
			require.NotNil(t, tempDir)
			defer cleanUpFn()

			// Create gno package
			createGnoModPkg(t, tempDir, pkgDir, tc.modData)

			_, err := ParseGnoMod(filepath.Join(tempDir, tc.modPath))
			if tc.errShouldContain != "" {
				assert.ErrorContains(t, err, tc.errShouldContain)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
