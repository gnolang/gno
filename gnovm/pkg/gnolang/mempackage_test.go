package gnolang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemPackage_Validate(t *testing.T) {
	t.Parallel()
	heyPackageFiles := []*std.MemFile{{Name: "a.gno", Body: "package hey"}}
	tt := []struct {
		name        string
		mpkg        *std.MemPackage
		errContains string
		panicMsg    string
	}{
		{
			"correct",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/r/hey",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"unsorted",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/r/path/to/pkg",
				Files: []*std.MemFile{{Name: "b.gno", Body: "package hey"}, {Name: "a.gno", Body: "package hey"}},
			},
			"unsorted",
			"",
		},
		{
			"duplicate",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/r/path/to/pkg",
				Files: []*std.MemFile{{Name: "a.gno", Body: "package hey"}, {Name: "a.gno", Body: "package hey"}},
			},
			"duplicate",
			"",
		},
		{
			"valid_long_path",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/r/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"invalid_path_length",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path\"",
		},
		{
			"invalid_path",
			&std.MemPackage{
				Type: MPUserProd,
				Name: "hey",
				// user package path for MPUserProd is more restricted. It starts with singl letter
				Path:  "example.com/path/def",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/path/def\"",
		},
		{
			"invalid_empty_path",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/r/path//def",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/r/path//def\"",
		},
		{
			"invalid_trailing_slash",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/p/path/abc/def/",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/p/path/abc/def/\"",
		},
		{
			"invalid_uppercase",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/PaTh/abc/def",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/PaTh/abc/def\"",
		},
		{
			"invalid_number",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/1Path/abc/def",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/1Path/abc/def\"",
		},

		{
			"special_character",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/p@th/abc/def",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/p@th/abc/def\"",
		},

		{
			"special_character_2",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "example.com/p&th/abc/def",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"example.com/p&th/abc/def\"",
		},
		{
			"leading_hyphen",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/-path",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/-path\"",
		},
		{
			"trailing_hyphen",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/path-",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/path-\"",
		},
		{
			"between_hyphen",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/p-ath",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"invalid_hyphen_1",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/-",
				Files: heyPackageFiles,
			},
			"invalid package/realm path",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/-\"",
		},
		{
			"invalid_hyphen_2",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/-/-",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/-/-\"",
		},
		{
			"invalid_hyphen_3",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/--/path",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/--/path\"",
		},
		{
			"leading_underscore",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/_path",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/_path\"",
		},
		{
			"trailing_underscore",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/path_",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/path_\"",
		},
		{
			"between_underscore",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/p_ath",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"invalid_underscore_1",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/_",
				Files: heyPackageFiles,
			},
			"invalid package/realm path",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/_\"",
		},
		{
			"invalid_underscore_2",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/_/_",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/_/_\"",
		},
		{
			"invalid_underscore_3",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/__/path",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/__/path\"",
		},
		{
			"consecutive_hyphen_in_segment",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/alice--bob",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/alice--bob\"",
		},
		{
			"consecutive_underscore_in_segment",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/alice__bob",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/alice__bob\"",
		},
		{
			"mixed_consecutive_separators_in_segment",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/alice-_bob",
				Files: heyPackageFiles,
			},
			"",
			"expected user package path for \"MPUserProd\" but got \"gno.land/r/path/alice-_bob\"",
		},
		{
			"futureproof_x", // XXX: we currently accept mempackages with any single-letter path, meaning that we need another layer of validation later.
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/x/path/path",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"custom_domain",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "github.com/p/path/path",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"valid_p_path",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/p/path/path",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"valid_r_path",
			&std.MemPackage{
				Type:  MPUserProd,
				Name:  "hey",
				Path:  "gno.land/r/path/path",
				Files: heyPackageFiles,
			},
			"",
			"",
		},
		{
			"valid_with_gno_toml",
			&std.MemPackage{
				Type: MPUserProd,
				Name: "hey",
				Path: "gno.land/r/path/path",
				Files: []*std.MemFile{
					{Name: "a.gno", Body: "package hey"},
					{Name: "bar.toml"},
					{Name: "foo.toml"},
				},
			},
			"",
			"",
		},
		{
			"valid_with_gno_toml_and_readme",
			&std.MemPackage{
				Type: MPUserProd,
				Name: "hey",
				Path: "gno.land/r/path/path",
				Files: []*std.MemFile{
					{Name: "README.md", Body: "# Hey Package"},
					{Name: "a.gno", Body: "package hey"},
					{Name: "foo.toml"},
				},
			},
			"",
			"",
		},
		{
			"valid_with_other_markdown",
			&std.MemPackage{
				Type: MPUserProd,
				Name: "hey",
				Path: "gno.land/r/path/path",
				Files: []*std.MemFile{
					{Name: "a.gno", Body: "package hey"},
					{Name: "other.md", Body: "# Other markdown file"},
				},
			},
			"",
			"",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.panicMsg != "" {
				assert.PanicsWithValue(t, tc.panicMsg, func() {
					_ = ValidateMemPackage(tc.mpkg)
				})
			} else {
				assert.NotPanics(t, func() {
					err := ValidateMemPackage(tc.mpkg)
					if tc.errContains == "" {
						assert.NoError(t, err)
					} else {
						assert.ErrorContains(t, err, tc.errContains)
					}
				})
			}
		})
	}
}

// TestValidateMemPackageAny_FlatName asserts ValidateMemPackageAny rejects
// any MemFile.Name containing a slash — subdirs are a write-routing convention
// only, not part of the in-memory name.
func TestValidateMemPackageAny_FlatName(t *testing.T) {
	t.Parallel()
	tt := []struct {
		fname       string
		errContains string // empty == should validate
	}{
		// allowed (a base file `main.gno` is always present)
		{"foo.gno", ""},
		{"foo_filetest.gno", ""},

		// rejected by the file-name regex via MemFile.ValidateBasic
		{"filetests/foo.gno", "invalid file name"},
		{"filetests/foo_filetest.gno", "invalid file name"},
		{"sub/foo.gno", "invalid file name"},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.fname, func(t *testing.T) {
			t.Parallel()
			mpkg := &std.MemPackage{
				Type: MPUserAll, // accepts both prod and filetest files
				Name: "hey",
				Path: "example.com/r/hey",
				Files: []*std.MemFile{
					{Name: "main.gno", Body: "package hey\n"},
					{Name: tc.fname, Body: "package hey\n"},
				},
			}
			mpkg.Sort()
			err := ValidateMemPackageAny(mpkg)
			if tc.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

// TestIsTestFile_Suffixes asserts IsTestFile recognizes both `_test.gno` and
// `_filetest.gno` suffixes. MemFile.Name is flat — never a `filetests/` path.
func TestIsTestFile_Suffixes(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name string
		want bool
	}{
		{"foo.gno", false},
		{"foo_test.gno", true},
		{"foo_filetest.gno", true},
		{"README.md", false},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsTestFile(tc.name))
		})
	}
}

// TestReadMemPackage_FiletestRouting exercises the on-disk → MemPackage
// pipeline: any `.gno` file under the filetests/ subdir is loaded with its
// bare basename as MemFile.Name and Kind=KindFiletest (the subdir IS the
// classification — no `_filetest.gno` suffix required). Round-trip via WriteTo
// lands every filetest back under filetests/.
func TestReadMemPackage_FiletestRouting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.gno"), []byte("package hey\n"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "filetests"), 0o755))
	// New-style filetest: bare basename, no `_filetest.gno` suffix.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "filetests", "new.gno"), []byte("package hey\n"), 0o644))
	// Legacy filetest at root, still recognized via suffix.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "legacy_filetest.gno"), []byte("package hey\n"), 0o644))

	mpkg, err := ReadMemPackage(dir, "example.com/r/hey", MPUserAll)
	require.NoError(t, err)

	byName := make(map[string]*std.MemFile, len(mpkg.Files))
	for _, f := range mpkg.Files {
		byName[f.Name] = f
	}
	assert.NotNil(t, byName["a.gno"], "expected a.gno at root")
	assert.NotNil(t, byName["new.gno"], "filetest loaded with bare basename")
	assert.Nil(t, byName["filetests/new.gno"], "MemFile.Name must not carry the filetests/ prefix")
	assert.NotNil(t, byName["legacy_filetest.gno"], "legacy suffix at root still loaded")
	// Kind is stamped from disk location (or suffix for legacy at root).
	assert.Equal(t, std.KindFiletest, byName["new.gno"].Kind)
	assert.Equal(t, std.KindFiletest, byName["legacy_filetest.gno"].Kind)

	// Round-trip: write and re-read; WriteTo routes filetests under filetests/
	// regardless of suffix, classifying by Kind (with legacy-suffix fallback).
	out := t.TempDir()
	require.NoError(t, mpkg.WriteTo(out))
	_, err = os.Stat(filepath.Join(out, "filetests", "new.gno"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(out, "filetests", "legacy_filetest.gno"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(out, "a.gno"))
	require.NoError(t, err)
}
