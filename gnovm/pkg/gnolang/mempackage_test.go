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
	}{
		{
			"correct",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/r/hey",
				Files: heyPackageFiles,
			},
			"",
		},
		{
			"unsorted",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/path/to/pkg",
				Files: []*std.MemFile{{Name: "b.gno", Body: "package hey"}, {Name: "a.gno", Body: "package hey"}},
			},
			"unsorted",
		},
		{
			"duplicate",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/path/to/pkg",
				Files: []*std.MemFile{{Name: "a.gno", Body: "package hey"}, {Name: "a.gno", Body: "package hey"}},
			},
			"duplicate",
		},
		{
			"invalid_path_length",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/very/long/path",
				Files: heyPackageFiles,
			},
			"path length",
		},
		{
			"invalid_empty_path",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/path//def",
				Files: heyPackageFiles,
			},
			"invalid package path",
		},
		{
			"invalid_trailing_slash",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/path/abc/def/",
				Files: heyPackageFiles,
			},
			"invalid package path",
		},
		{
			"invalid_uppercase",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/PaTh/abc/def",
				Files: heyPackageFiles,
			},
			"invalid package path",
		},
		{
			"invalid_number",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/1Path/abc/def",
				Files: heyPackageFiles,
			},
			"invalid package path",
		},
		{
			"special_character",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/p@th/abc/def",
				Files: heyPackageFiles,
			},
			"invalid package path",
		},
		{
			"special_character_2",
			&std.MemPackage{
				Name:  "hey",
				Path:  "example.com/p&th/abc/def",
				Files: heyPackageFiles,
			},
			"invalid package path",
		},
		{
			"leading_underscore",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_path",
				Files: heyPackageFiles,
			},
			"",
		},
		{
			"trailing_underscore",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path_",
				Files: heyPackageFiles,
			},
			"",
		},
		{
			"between_underscore",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/p_ath",
				Files: heyPackageFiles,
			},
			"",
		},
		{
			"invalid_underscore_1",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_",
				Files: heyPackageFiles,
			},
			"invalid package/realm path",
		},
		{
			"invalid_underscore_2",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/_/_",
				Files: heyPackageFiles,
			},
			"invalid package/realm path",
		},
		{
			"invalid_underscore_3",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/__/path",
				Files: heyPackageFiles,
			},
			"invalid package/realm path",
		},
		{
			"futureproof_x", // XXX: we currently accept mempackages with any single-letter path, meaning that we need another layer of validation later.
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/x/path/path",
				Files: heyPackageFiles,
			},
			"",
		},
		{
			"custom_domain",
			&std.MemPackage{
				Name:  "hey",
				Path:  "github.com/p/path/path",
				Files: heyPackageFiles,
			},
			"",
		},
		{
			"valid_p_path",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/p/path/path",
				Files: heyPackageFiles,
			},
			"",
		},
		{
			"valid_r_path",
			&std.MemPackage{
				Name:  "hey",
				Path:  "gno.land/r/path/path",
				Files: heyPackageFiles,
			},
			"",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateMemPackage(tc.mpkg)
			if tc.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
