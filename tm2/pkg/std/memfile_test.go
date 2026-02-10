package std

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
