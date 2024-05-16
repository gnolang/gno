package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemPackage_Validate(t *testing.T) {
	tt := []struct {
		name           string
		mpkg           *MemPackage
		errMayContains string
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
			`mempackage "gno.land/r/demo/hey" has unsorted files`,
		},
		{
			"Duplicate",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "a.gno"}, {Name: "a.gno"}},
			},
			`duplicate file name "a.gno"`,
		},
		{
			"InvalidPathLength",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			`invalid length of package/realm path`,
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
			"error",
		},
		{
			"Invalid underscore 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_/_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid underscore 3",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/__/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid hyphen",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/pa-th",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid x",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/x/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid missing path 1",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid missing path 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid path",
			&MemPackage{
				Name:  "hey",
				Path:  "github.com/p/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p@th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Special character 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p&th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid number",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/1Path/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid uppercase",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/PaTh/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid empty path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path//def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/abc/def/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
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
			"error",
		},
		{
			"Invalid long path with trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
		{
			"Invalid long path with empty",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very//long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			"error",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.mpkg.Validate()
			if tc.errMayContains == "" {
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
