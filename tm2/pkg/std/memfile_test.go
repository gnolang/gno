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
			"Correct",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"",
		},
		{
			"Unsorted",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "b.gno"}, {Name: "a.gno"}},
			},
			"unsorted",
		},
		{
			"Duplicate",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "a.gno"}, {Name: "a.gno"}},
			},
			"duplicate",
		},
		{
			"InvalidPathLength",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"path length",
		},
		{
			"valid p",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"",
		},
		{
			"valid r",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"",
		},
		{
			"Leading underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"",
		},
		{
			"Trailing underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"",
		},
		{
			"Between underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/p_ath",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"",
		},
		{
			"Invalid underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid underscore 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_/_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid underscore 3",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/__/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid hyphen",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/pa-th",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid x",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/x/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid missing path 1",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid missing path 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid path",
			&MemPackage{
				Name:  "hey",
				Path:  "github.com/p/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p@th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Special character 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p&th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid number",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/1Path/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid uppercase",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/PaTh/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid empty path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path//def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/abc/def/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"valid long path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"",
		},
		{
			"Invalid long path with special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/p@th",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid long path with trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
		{
			"Invalid long path with empty",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very//long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"invalid package/realm path",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.mpkg.Validate()
			if tc.errContains != "" {
				assert.ErrorContains(t, err, tc.errContains)
			} else {
				assert.NoError(t, err)
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
		t.Run(tt.name, func(t *testing.T) {
			dirPath, filename := SplitFilepath(tt.filepath)
			assert.Equal(t, tt.expDirPath, dirPath)
			assert.Equal(t, tt.expFilename, filename)
		})
	}
}
