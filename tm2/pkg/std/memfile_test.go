package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemPackage_Validate(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name          string
		mpkg          *MemPackage
		shouldHaveErr bool
		errContains   string
	}{
		{
			"Correct",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
			"",
		},
		{
			"Unsorted",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "b.gno"}, {Name: "a.gno"}},
			},
			true,
			"unsorted",
		},
		{
			"Duplicate",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/demo/hey",
				Files: []*MemFile{{Name: "a.gno"}, {Name: "a.gno"}},
			},
			true,
			"duplicate",
		},
		{
			"InvalidPathLength",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"path length",
		},
		{
			"valid p",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
			"",
		},
		{
			"valid r",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
			"",
		},
		{
			"Leading underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
			"",
		},
		{
			"Trailing underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
			"",
		},
		{
			"Between underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/p_ath",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
			"",
		},
		{
			"Invalid underscore",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid underscore 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_/_",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid underscore 3",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/__/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid hyphen",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/pa-th",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid x",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/x/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid missing path 1",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid missing path 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid path",
			&MemPackage{
				Name:  "hey",
				Path:  "github.com/p/path/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p@th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Special character 2",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/p&th/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid number",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/1Path/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid uppercase",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/PaTh/abc/def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid empty path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path//def",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/abc/def/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"valid long path",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			false,
			"",
		},
		{
			"Invalid long path with special character",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/p@th",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid long path with trailing slash",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very/long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
		{
			"Invalid long path with empty",
			&MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/very/very/very//long/path/",
				Files: []*MemFile{{Name: "a.gno"}},
			},
			true,
			"invalid package/realm path",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			err := tc.mpkg.Validate()
			if tc.shouldHaveErr {
				assert.ErrorContains(t, err, tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
