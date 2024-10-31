package gnovm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemPackage_Validate(t *testing.T) {
	fileA := &MemFile{
		Name: "a.gno",
		Body: "package test",
	}
	fileB := &MemFile{
		Name: "b.gno",
		Body: "package test",
	}

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
				Files: []*MemFile{fileA},
			},
			"",
		},
		{
			"Unsorted",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{fileB, fileA},
			},
			"unsorted",
		},
		{
			"Duplicate",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{fileA, fileA},
			},
			"duplicate",
		},
		{
			"InvalidPathLength",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: []*MemFile{fileA},
			},
			"path length",
		},
		{
			"valid p",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/path",
				Files: []*MemFile{fileA},
			},
			"",
		},
		{
			"valid r",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path",
				Files: []*MemFile{fileA},
			},
			"",
		},
		{
			"Leading underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_path",
				Files: []*MemFile{fileA},
			},
			"",
		},
		{
			"Trailing underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path_",
				Files: []*MemFile{fileA},
			},
			"",
		},
		{
			"Between underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/p_ath",
				Files: []*MemFile{fileA},
			},
			"",
		},
		{
			"Invalid underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid underscore 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_/_",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid underscore 3",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/__/path",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid hyphen",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/pa-th",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid x",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/x/path/path",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid missing path 1",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid missing path 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid path",
			&MemPackage{
				Name:  "hey",
				Path:  "github.com/p/path/path",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p@th/abc/def",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Special character 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p&th/abc/def",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid number",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/1Path/abc/def",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid uppercase",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/PaTh/abc/def",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid empty path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path//def",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/abc/def/",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"valid long path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path",
				Files: []*MemFile{fileA},
			},
			"",
		},
		{
			"Invalid long path with special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/p@th",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid long path with trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path/",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid long path with empty",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very//long/path/",
				Files: []*MemFile{fileA},
			},
			"invalid package/realm path",
		},
		{
			"Invalid package imports realm",
			&MemPackage{
				Name: "test",
				Path: "gno.land/p/demo/test",
				Files: []*MemFile{
					{Name: "a.gno", Body: `
					package test

					import "gno.land/r/demo/avl"
					
					func A() {
						avl.A()
					}
					`},
				},
			},
			"package \"gno.land/p/demo/test\" imports realm \"gno.land/r/demo/avl\"",
		},
		{
			"Valid witr /r/ as a realm name",
			&MemPackage{
				Name: "test",
				Path: "gno.land/p/demo/test",
				Files: []*MemFile{
					{Name: "a.gno", Body: `
					package test

					import "gno.land/p/r/r"
					
					func A() {
						r.A()
					}
					`},
				},
			},
			"",
		},
		{
			"Valid package containing non gno file",
			&MemPackage{
				Name: "test",
				Path: "gno.land/p/demo/test",
				Files: []*MemFile{
					{
						Name: "README.md",
						Body: `
						# Test
						`,
					},
					{Name: "a.gno", Body: `
					package test

					import "gno.land/p/r/r"
					
					func A() {
						r.A()
					}
					`},
				},
			},
			"",
		},
		{
			"Invalid empty gno file",
			&MemPackage{
				Name: "test",
				Path: "gno.land/p/demo/test",
				Files: []*MemFile{
					{Name: "a.gno"},
				},
			},
			"failed to parse imports in file \"a.gno\" of package \"gno.land/p/demo/test\"",
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dirPath, filename := SplitFilepath(tt.filepath)
			assert.Equal(t, tt.expDirPath, dirPath)
			assert.Equal(t, tt.expFilename, filename)
		})
	}
}
