package std

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemPackage_Validate(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name        string
		mpkg        *MemPackage
		errContains string
	}{
		{
			"correct",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/path/to/pkg",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"",
		},
		{
			"unsorted",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/path/to/pkg",
				Files: []*MemFile{{Name: "b.txt"}, {Name: "a.txt"}},
			},
			"unsorted",
		},
		{
			"duplicate",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/path/to/pkg",
				Files: []*MemFile{{Name: "a.txt"}, {Name: "a.txt"}},
			},
			"duplicate",
		},
		{
			"invalid_path_length",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"path length",
		},
		{
			"invalid_empty_path",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/path//def",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"invalid package path",
		},
		{
			"invalid_trailing_slash",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/path/abc/def/",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"invalid package path",
		},
		{
			"invalid_uppercase",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/PaTh/abc/def",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"invalid package path",
		},
		{
			"invalid_number",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/1Path/abc/def",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"invalid package path",
		},
		{
			"special_character",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/p@th/abc/def",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"invalid package path",
		},
		{
			"special_character_2",
			&MemPackage{
				Name:  "hey",
				Path:  "example.com/p&th/abc/def",
				Files: []*MemFile{{Name: "a.txt"}},
			},
			"invalid package path",
		},
		{
			"Valid toml file",
			&MemPackage{
				Name: "hey",
				Path: "gno.land/r/demo/hey",
				Files: []*MemFile{
					{Name: "a.gno"},
					{Name: "gnomod.toml"},
				},
			},
			"",
		},
		{
			"Multiple toml files",
			&MemPackage{
				Name: "hey",
				Path: "gno.land/r/demo/hey",
				Files: []*MemFile{
					{Name: "a.gno"},
					{Name: "gnomod.toml"},
					{Name: "gnoweb.toml"},
				},
			},
			"",
		},
		{
			"Toml file without gno file",
			&MemPackage{
				Name: "hey",
				Path: "gno.land/r/demo/hey",
				Files: []*MemFile{
					{Name: "gnomod.toml"},
				},
			},
			"",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.mpkg.ValidateBasic()
			if tc.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

func TestIsFiletestName(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name string
		want bool
	}{
		// new-style: filetests/ + .gno
		{"filetests/foo.gno", true},
		{"filetests/foo_filetest.gno", true},
		{"filetests/a.b.gno", true},

		// non-.gno under filetests/ is not a filetest
		{"filetests/README.md", false},
		{"filetests/gno.mod", false},
		{"filetests/foo.toml", false},

		// legacy suffix at the root still counts
		{"foo_filetest.gno", true},

		// regular files
		{"foo.gno", false},
		{"foo_test.gno", false},
		{"README.md", false},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsFiletestName(tc.name))
		})
	}
}

func TestMemFile_ValidateBasic_FiletestsDir(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name  string
		valid bool
	}{
		// allowed
		{"foo.gno", true},
		{"filetests/foo.gno", true},
		{"filetests/foo_filetest.gno", true},
		{"filetests/a.b.gno", true},
		{"README.md", true},
		{"LICENSE", true},

		// rejected
		{"filetests/README.md", false},   // non-.gno under filetests/
		{"filetests/foo.toml", false},    // non-.gno under filetests/
		{"filetests/sub/foo.gno", false}, // nested subdir
		{"sub/foo.gno", false},           // other subdir
		{"filetests/", false},            // bare prefix
		{"/foo.gno", false},              // leading slash
		{"foo.gno/", false},              // trailing slash
		{"../foo.gno", false},            // traversal
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := (&MemFile{Name: tc.name, Body: "x"}).ValidateBasic()
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMemPackage_WriteTo_FiletestsLayout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mpkg := &MemPackage{
		Name: "hey",
		Path: "gno.land/r/demo/hey",
		Files: []*MemFile{
			{Name: "a.gno", Body: "package hey\n"},
			{Name: "filetests/new.gno", Body: "package hey\n"},
			// legacy: bare basename, no prefix in Name
			{Name: "legacy_filetest.gno", Body: "package hey\n"},
		},
	}
	require.NoError(t, mpkg.WriteTo(dir))

	// New-style: written under filetests/ as encoded in Name.
	body, err := os.ReadFile(filepath.Join(dir, "filetests", "new.gno"))
	require.NoError(t, err)
	assert.Equal(t, "package hey\n", string(body))

	// Legacy fallback: bare *_filetest.gno also routed under filetests/.
	body, err = os.ReadFile(filepath.Join(dir, "filetests", "legacy_filetest.gno"))
	require.NoError(t, err)
	assert.Equal(t, "package hey\n", string(body))

	// Root file stays at the root.
	body, err = os.ReadFile(filepath.Join(dir, "a.gno"))
	require.NoError(t, err)
	assert.Equal(t, "package hey\n", string(body))
}

func TestSplitFilepath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		filepath    string
		expDirPath  string
		expFilename string
	}{
		{
			name: "empty",
		},
		{
			name:       "one part",
			filepath:   "root",
			expDirPath: "root",
		},
		{
			name:        "file",
			filepath:    "gno.land/r/demo/avl/avl.gno",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "avl.gno",
		},
		{
			name:       "trailing slash",
			filepath:   "gno.land/r/demo/avl/",
			expDirPath: "gno.land/r/demo/avl",
		},
		{
			name:        "license",
			filepath:    "gno.land/r/demo/avl/LICENSE",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "LICENSE",
		},
		{
			name:        "readme",
			filepath:    "gno.land/r/demo/avl/README",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "README",
		},
		{
			name:       "regular path",
			filepath:   "gno.land/p/demo/avl",
			expDirPath: "gno.land/p/demo/avl",
		},
		{
			name:        "nested path with multiple files",
			filepath:    "gno.land/r/demo/avl/nested/file.gno",
			expDirPath:  "gno.land/r/demo/avl/nested",
			expFilename: "file.gno",
		},
		{
			name:        "path with dots",
			filepath:    "gno.land/r/demo/avl.test/file.gno",
			expDirPath:  "gno.land/r/demo/avl.test",
			expFilename: "file.gno",
		},
		{
			name:        "path with underscores",
			filepath:    "gno.land/r/demo/avl_test/file.gno",
			expDirPath:  "gno.land/r/demo/avl_test",
			expFilename: "file.gno",
		},
		{
			name:        "path with numbers",
			filepath:    "gno.land/r/demo/avl123/file.gno",
			expDirPath:  "gno.land/r/demo/avl123",
			expFilename: "file.gno",
		},
		{
			name:        "path with mixed case",
			filepath:    "gno.land/r/demo/avlTest/file.gno",
			expDirPath:  "gno.land/r/demo/avlTest",
			expFilename: "file.gno",
		},
		{
			name:        "path with special files",
			filepath:    "gno.land/r/demo/avl/.gitignore",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: ".gitignore",
		},
		{
			name:        "path with hidden directory",
			filepath:    "gno.land/r/demo/.avl/file.gno",
			expDirPath:  "gno.land/r/demo/.avl",
			expFilename: "file.gno",
		},
		{
			name:        "path with toml file",
			filepath:    "gno.land/r/demo/avl/config.toml",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "config.toml",
		},
		{
			name:        "path with markdown file",
			filepath:    "gno.land/r/demo/avl/README.md",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "README.md",
		},
		{
			name:        "path with json file",
			filepath:    "gno.land/r/demo/avl/config.json",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "config.json",
		},
		{
			name:        "path with yaml file",
			filepath:    "gno.land/r/demo/avl/config.yaml",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "config.yaml",
		},
		{
			name:        "path with multiple extensions",
			filepath:    "gno.land/r/demo/avl/file.test.gno",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "file.test.gno",
		},
		{
			name:        "path with spaces",
			filepath:    "gno.land/r/demo/avl test/file.gno",
			expDirPath:  "gno.land/r/demo/avl test",
			expFilename: "file.gno",
		},
		{
			name:        "path with unicode",
			filepath:    "gno.land/r/demo/avl-测试/file.gno",
			expDirPath:  "gno.land/r/demo/avl-测试",
			expFilename: "file.gno",
		},
		{
			name:        "path with emoji",
			filepath:    "gno.land/r/demo/avl-🚀/file.gno",
			expDirPath:  "gno.land/r/demo/avl-🚀",
			expFilename: "file.gno",
		},
		{
			name:        "path with multiple slashes",
			filepath:    "gno.land/r/demo/avl//file.gno",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "file.gno",
		},
		{
			name:        "path with leading slash",
			filepath:    "/gno.land/r/demo/avl/file.gno",
			expDirPath:  "/gno.land/r/demo/avl",
			expFilename: "file.gno",
		},
		{
			name:        "path with trailing slash and file",
			filepath:    "gno.land/r/demo/avl/file.gno/",
			expDirPath:  "gno.land/r/demo/avl/file.gno",
			expFilename: "",
		},
		{
			name:        "path with just filename",
			filepath:    "file.gno",
			expDirPath:  "file.gno",
			expFilename: "",
		},
		{
			name:        "path with just directory",
			filepath:    "gno.land/r/demo/avl/",
			expDirPath:  "gno.land/r/demo/avl",
			expFilename: "",
		},
		{
			name:        "path with single character",
			filepath:    "gno.land/r/demo/a/file.gno",
			expDirPath:  "gno.land/r/demo/a",
			expFilename: "file.gno",
		},
		{
			name:        "path with maximum length",
			filepath:    "gno.land/r/demo/" + strings.Repeat("a", 100) + "/file.gno",
			expDirPath:  "gno.land/r/demo/" + strings.Repeat("a", 100),
			expFilename: "file.gno",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dirPath, filename := SplitFilepath(tt.filepath)
			assert.Equal(t, tt.expDirPath, dirPath)
			assert.Equal(t, tt.expFilename, filename)
		})
	}
}
