package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemPackage_Validate(t *testing.T) {
	tt := []struct {
		name           string
		mpkg           *MemPackage
		shouldHaveErr bool
	}{
		{
			"Correct",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
		},
		{
			"Unsorted",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "b.gno"}, {Name: "a.gno"}},
			},
			true,
		},
		{
			"Duplicate",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "a.gno"}, {Name: "a.gno"}},
			},
			true,
		},
		{
			"InvalidPathLength",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"valid p",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
		},
		{
			"valid r",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
		},
		{
			"Leading underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
		},
		{
			"Trailing underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
		},
		{
			"Between underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/p_ath",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
		},
		{
			"Invalid underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid underscore 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_/_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid underscore 3",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/__/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid hyphen",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/pa-th",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid x",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/x/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid missing path 1",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid missing path 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid path",
			&MemPackage{
				Name:  "hey",
				Path:  "github.com/p/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p@th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Special character 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p&th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid number",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/1Path/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid uppercase",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/PaTh/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid empty path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path//def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/abc/def/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"valid long path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
		},
		{
			"Invalid long path with special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/p@th",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid long path with trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
		{
			"Invalid long path with empty",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very//long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.mpkg.Validate()
			if tc.shouldHaveErr {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
