package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
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
				Path:  "example.com/r/path/to/hey",
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
				Path:  "example.com/r/path/to/hey",
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
				Path:  "example.com/r/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/hey",
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
				Path:  "gno.land/r/p_ath/hey",
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
				Path:  "gno.land/x/path/hey",
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
				Path:  "github.com/p/path/hey",
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
				Path:  "gno.land/p/path/hey",
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
				Path:  "gno.land/r/path/hey",
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
				Path: "gno.land/r/path/hey",
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
				Path: "gno.land/r/path/hey",
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
				Path: "gno.land/r/path/hey",
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

// '#' is reserved for sub-realm pkgpath synthesis (realm.Sub); the
// reservation is an explicit check with its own diagnostic, ahead of
// the path regexes that also exclude it.
func TestValidateMemPackageAny_SepReserved(t *testing.T) {
	t.Parallel()
	mpkg := &std.MemPackage{
		Type:  MPUserProd,
		Name:  "hey",
		Path:  "example.com/r/hey#sub",
		Files: []*std.MemFile{{Name: "a.gno", Body: "package hey"}},
	}
	err := ValidateMemPackageAny(mpkg)
	assert.ErrorContains(t, err, "is reserved for sub-realm derivation")
}

func TestIsVersionSuffix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"v1", true}, // v1 is allowed for backwards compatibility
		{"v2", true},
		{"v3", true},
		{"v10", true},
		{"v99", true},
		{"v0", true}, // v0 is a valid version suffix (e.g. gno.land/p/nt/avl/v0)
		{"v", false}, // incomplete
		{"v2beta", false},
		{"2", false}, // missing v prefix
		{"", false},
		{"V2", false}, // uppercase not allowed
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := isVersionSuffix(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestLastPathElement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path     string
		expected string
	}{
		{"gno.land/r/demo/counter", "counter"},
		{"gno.land/p/demo/avl", "avl"},
		{"gno.land/r/demo/foo/v1", "foo"},  // v1 is version suffix, return foo
		{"gno.land/r/demo/foo/v2", "foo"},  // v2 is version suffix, return foo
		{"gno.land/r/demo/foo/v10", "foo"}, // multi-digit version
		{"encoding/json", "json"},          // stdlib
		{"gno.land/r/demo/v2app", "v2app"}, // v2 in name, not suffix
		{"gno.land/r/demo/myv2", "myv2"},   // v2 suffix in name
		{"single", "single"},               // single element path
		{"", ""},                           // empty path
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			got := LastPathElement(tt.path)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestValidatePkgNameMatchesPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		pkgName     Name
		pkgPath     string
		errContains string
	}{
		{"match", "counter", "gno.land/r/demo/counter", ""},
		{"mismatch", "hello", "gno.land/r/demo/goodbye", "does not match path element"},
		{"version_suffix_v1_match", "foo", "gno.land/r/demo/foo/v1", ""},
		{"version_suffix_v1_mismatch", "bar", "gno.land/r/demo/foo/v1", "does not match path element"},
		{"version_suffix_v2_match", "foo", "gno.land/r/demo/foo/v2", ""},
		{"version_suffix_v2_mismatch", "bar", "gno.land/r/demo/foo/v2", "does not match path element"},
		{"stdlib_match", "json", "encoding/json", ""},
		{"stdlib_mismatch", "xml", "encoding/json", "does not match path element"},
		{"empty_path", "foo", "", ""},
		{"consecutive_version_suffixes", "v2", "gno.land/r/demo/foo/v2/v3", "consecutive version suffixes"},
		{"consecutive_version_suffixes_any_name", "foo", "gno.land/r/demo/foo/v1/v2", "consecutive version suffixes"},
		{"consecutive_version_suffixes_only", "v2", "v1/v2", "consecutive version suffixes"},
		{"version_suffix_under_namespace", "demo", "gno.land/r/demo/v2", ""},
		{"version_suffix_single_element", "v2", "v2", ""}, // degenerate, not consecutive
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePkgNameMatchesPath(tt.pkgName, tt.pkgPath)
			if tt.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errContains)
			}
		})
	}
}
