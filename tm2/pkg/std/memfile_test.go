package std

import (
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
			name:       "regular path",
			filepath:   "gno.land/p/demo/avl",
			expDirPath: "gno.land/p/demo/avl",
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
