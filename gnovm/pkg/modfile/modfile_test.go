package modfile_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/modfile"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModfileStructTags(t *testing.T) {
	t.Parallel()
	mf := modfile.Modfile{
		PkgPath:  "gno.land/p/demo/testtags",
		Draft:    true,
		Uploader: "gno.land/r/testuploader",
	}

	// Test TOML marshaling
	tomlBytes, err := toml.Marshal(mf)
	require.NoError(t, err)
	tomlString := string(tomlBytes)
	assert.Contains(t, tomlString, `pkgPath = "gno.land/p/demo/testtags"`)
	assert.Contains(t, tomlString, `draft = true`)
	assert.Contains(t, tomlString, `uploader = "gno.land/r/testuploader"`)

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(mf)
	require.NoError(t, err)
	jsonString := string(jsonBytes)
	assert.Contains(t, jsonString, `"pkgPath":"gno.land/p/demo/testtags"`)
	assert.Contains(t, jsonString, `"draft":true`)
	assert.Contains(t, jsonString, `"uploader":"gno.land/r/testuploader"`)

	// Test omitempty for Draft and Uploader (when false/empty)
	mfOmitempty := modfile.Modfile{
		PkgPath: "gno.land/p/demo/omitempty",
	}
	tomlBytesOmitempty, err := toml.Marshal(mfOmitempty)
	require.NoError(t, err)
	tomlStringOmitempty := string(tomlBytesOmitempty)
	assert.NotContains(t, tomlStringOmitempty, "draft =")
	assert.NotContains(t, tomlStringOmitempty, "uploader =")

	jsonBytesOmitempty, err := json.Marshal(mfOmitempty)
	require.NoError(t, err)
	jsonStringOmitempty := string(jsonBytesOmitempty)
	assert.NotContains(t, jsonStringOmitempty, `"draft":`)
	assert.NotContains(t, jsonStringOmitempty, `"uploader":`)
}

func TestParseModfile(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc             string
		inputData        string
		expectedModfile  *modfile.Modfile
		expectedErrorIs  error
		errShouldContain string
	}{
		{
			desc: "valid modfile content",
			inputData: `pkgPath = "gno.land/p/demoparse"
draft = true
uploader = "gno.land/r/test"`,
			expectedModfile: &modfile.Modfile{
				PkgPath:  "gno.land/p/demoparse",
				Draft:    true,
				Uploader: "gno.land/r/test",
			},
		},
		{
			desc:      "valid modfile content, only pkgPath",
			inputData: `pkgPath = "gno.land/p/demopkgpathonly"`,
			expectedModfile: &modfile.Modfile{
				PkgPath: "gno.land/p/demopkgpathonly",
			},
		},
		{
			desc:             "malformed TOML content",
			inputData:        `pkgPath = gno.land/p/malformed`, // Missing quotes
			errShouldContain: "failed to unmarshal gno.toml",
		},
		{
			desc:            "empty PkgPath",
			inputData:       `draft = true`, // PkgPath is required
			expectedErrorIs: modfile.ErrPkgPathEmpty,
		},
		{
			desc:            "empty input data",
			inputData:       "",
			expectedErrorIs: modfile.ErrPkgPathEmpty, // Unmarshals to empty struct, then validation fails
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			mf, err := modfile.ParseModfile([]byte(tc.inputData))
			if tc.expectedErrorIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErrorIs)
				if tc.errShouldContain != "" {
					assert.Contains(t, err.Error(), tc.errShouldContain)
				}
				assert.Nil(t, mf)
				return
			} else if tc.errShouldContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errShouldContain)
				assert.Nil(t, mf)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedModfile, mf)
		})
	}
}

func TestCreateModfile(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc             string
		modfileToCreate  *modfile.Modfile // The Modfile struct to attempt to create
		expectedContent  modfile.Modfile  // Expected if creation is successful (used for read-back verification)
		rootDirFn        func(t *testing.T) string
		preexistingToml  bool
		expectedErrorIs  error
		errShouldContain string
	}{
		{
			desc: "valid Modfile, new file in temp dir",
			modfileToCreate: &modfile.Modfile{
				PkgPath:  "gno.land/p/democreate",
				Draft:    false, // Explicitly set, omitempty means it won't be in file if false
				Uploader: "",    // Explicitly set, omitempty means it won't be in file if empty
			},
			expectedContent: modfile.Modfile{
				PkgPath: "gno.land/p/democreate",
				// Draft and Uploader will be zero-valued if not set or empty in marshalled TOML due to omitempty
			},
			rootDirFn: func(t *testing.T) string { return t.TempDir() },
		},
		{
			desc: "Modfile with empty PkgPath",
			modfileToCreate: &modfile.Modfile{
				PkgPath: "", // Empty PkgPath
				Draft:   true,
			},
			rootDirFn:       func(t *testing.T) string { return t.TempDir() },
			expectedErrorIs: modfile.ErrPkgPathEmpty,
		},
		{
			desc:             "nil Modfile provided",
			modfileToCreate:  nil,
			rootDirFn:        func(t *testing.T) string { return t.TempDir() },
			errShouldContain: "provided modfile data cannot be nil",
		},
		{
			desc: "non-absolute rootDir",
			modfileToCreate: &modfile.Modfile{
				PkgPath: "gno.land/p/nonabs",
			},
			rootDirFn:        func(t *testing.T) string { return "./testdir_nonabs" },
			errShouldContain: "is not absolute",
		},
		{
			desc: "gno.toml already exists",
			modfileToCreate: &modfile.Modfile{
				PkgPath: "gno.land/p/exists",
			},
			rootDirFn:       func(t *testing.T) string { return t.TempDir() },
			preexistingToml: true,
			expectedErrorIs: modfile.ErrModfileExists,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			rootDir := tc.rootDirFn(t)
			if !strings.Contains(rootDir, t.TempDir()) {
				// For non-temp dirs, ensure they are created for the test and cleaned up
				if !strings.Contains(tc.errShouldContain, "is not absolute") { // only if we don't expect this specific error
					errMkdir := os.MkdirAll(rootDir, 0755)
					require.NoError(t, errMkdir, "Failed to create rootDir for test: %s", rootDir)
				}
				defer os.RemoveAll(rootDir)
			}

			if tc.preexistingToml {
				// Ensure directory exists before creating pre-existing file, only if it's supposed to be absolute for this phase
				if filepath.IsAbs(rootDir) {
					errMkdir := os.MkdirAll(rootDir, 0755)
					require.NoError(t, errMkdir)
				}
				require.NoError(t, os.WriteFile(filepath.Join(rootDir, "gno.toml"), []byte("pre-existing content"), 0644))
			}

			err := modfile.CreateModfile(rootDir, tc.modfileToCreate)

			if tc.expectedErrorIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErrorIs)
				if tc.errShouldContain != "" { // Also check contains if specified (e.g. for wrapped errors)
					assert.Contains(t, err.Error(), tc.errShouldContain)
				}
				return
			} else if tc.errShouldContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errShouldContain)
				return
			}
			require.NoError(t, err)

			// Verify by reading back with ReadModfile (findInParents=false)
			createdMf, readErr := modfile.ReadModfile(rootDir, false)
			require.NoError(t, readErr, "Failed to read back created gno.toml")
			assert.Equal(t, tc.expectedContent, *createdMf)
		})
	}
}

func TestReadModfile(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T, baseDir string, tomlRelPath string, content *modfile.Modfile) string {
		targetDir := baseDir
		if tomlRelPath != "." && tomlRelPath != "" {
			targetDir = filepath.Join(baseDir, tomlRelPath)
		}
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		if content != nil {
			tomlBytes, err := toml.Marshal(*content)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(targetDir, "gno.toml"), tomlBytes, 0644)
			require.NoError(t, err)
		}
		return targetDir // returns the directory where the gno.toml was placed (or should have been)
	}

	testCases := []struct {
		desc             string
		tomlSetupPath    string // Relative path from baseDir to create gno.toml, or "." for baseDir itself
		searchDirRelPath string // Relative path from baseDir to start search, or "." for baseDir itself
		findInParents    bool
		modfileContent   *modfile.Modfile // Content to put in gno.toml, nil if no file should be created
		expectedModfile  *modfile.Modfile // Expected result, nil if error is expected
		errShouldContain string
	}{
		{
			desc:             "read from current dir, findInParents=false",
			tomlSetupPath:    ".",
			searchDirRelPath: ".",
			findInParents:    false,
			modfileContent:   &modfile.Modfile{PkgPath: "gno.land/p/current"},
			expectedModfile:  &modfile.Modfile{PkgPath: "gno.land/p/current"},
		},
		{
			desc:             "not found in current dir, findInParents=false",
			tomlSetupPath:    "sub", // gno.toml is in sub, but we search current
			searchDirRelPath: ".",
			findInParents:    false,
			modfileContent:   &modfile.Modfile{PkgPath: "gno.land/p/sub"},
			errShouldContain: modfile.ErrModfileNotFound.Error(),
		},
		{
			desc:             "read from current dir, findInParents=true",
			tomlSetupPath:    ".",
			searchDirRelPath: ".",
			findInParents:    true,
			modfileContent:   &modfile.Modfile{PkgPath: "gno.land/p/currenttrue"},
			expectedModfile:  &modfile.Modfile{PkgPath: "gno.land/p/currenttrue"},
		},
		{
			desc:             "read from parent dir, findInParents=true",
			tomlSetupPath:    ".", // gno.toml in baseDir (parent of searchDir)
			searchDirRelPath: "child",
			findInParents:    true,
			modfileContent:   &modfile.Modfile{PkgPath: "gno.land/p/parentsearch"},
			expectedModfile:  &modfile.Modfile{PkgPath: "gno.land/p/parentsearch"},
		},
		{
			desc:             "read from grandparent dir, findInParents=true",
			tomlSetupPath:    ".", // gno.toml in baseDir (grandparent of searchDir)
			searchDirRelPath: "parent/child",
			findInParents:    true,
			modfileContent:   &modfile.Modfile{PkgPath: "gno.land/p/grandparentsearch"},
			expectedModfile:  &modfile.Modfile{PkgPath: "gno.land/p/grandparentsearch"},
		},
		{
			desc:             "not found anywhere, findInParents=true",
			tomlSetupPath:    ".",
			searchDirRelPath: ".",
			findInParents:    true,
			modfileContent:   nil, // No gno.toml created
			errShouldContain: modfile.ErrModfileNotFound.Error(),
		},
		{
			desc:             "malformed toml, findInParents=true",
			tomlSetupPath:    ".",
			searchDirRelPath: ".",
			findInParents:    true,
			modfileContent:   &modfile.Modfile{PkgPath: "#invalid\npath"}, // Will cause marshal error or be invalid TOML
			// If Marshal itself fails, this test setup needs adjustment. For now, assume it produces bad TOML file for ParseModfile.
			// Correct way for bad TOML: Write raw malformed string directly.
			errShouldContain: "failed to unmarshal gno.toml",
		},
		{
			desc:             "PkgPath empty after parsing, findInParents=true",
			tomlSetupPath:    ".",
			searchDirRelPath: ".",
			findInParents:    true,
			modfileContent:   &modfile.Modfile{Draft: true}, // PkgPath is empty
			errShouldContain: "pkgPath must be set in gno.toml",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			baseDir := t.TempDir()

			// Special handling for malformed TOML test case to write raw bad data
			if strings.Contains(tc.desc, "malformed toml") {
				tomlDir := setup(t, baseDir, tc.tomlSetupPath, nil) // Setup directory structure without writing valid toml
				err := os.WriteFile(filepath.Join(tomlDir, "gno.toml"), []byte("pkgPath = not_a_string"), 0644)
				require.NoError(t, err)
			} else {
				_ = setup(t, baseDir, tc.tomlSetupPath, tc.modfileContent) // Normal setup
			}

			searchDir := baseDir
			if tc.searchDirRelPath != "." && tc.searchDirRelPath != "" {
				searchDir = filepath.Join(baseDir, tc.searchDirRelPath)
				// Ensure searchDir itself exists, especially if it's deeper than where gno.toml might be
				err := os.MkdirAll(searchDir, 0755)
				require.NoError(t, err)
			}

			actualMf, err := modfile.ReadModfile(searchDir, tc.findInParents)

			if tc.errShouldContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errShouldContain)
				if tc.errShouldContain == modfile.ErrModfileNotFound.Error() {
					assert.ErrorIs(t, err, modfile.ErrModfileNotFound)
				}
				assert.Nil(t, actualMf)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedModfile, actualMf)
		})
	}
}
