package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemPackage_Validate(t *testing.T) {
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
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.mpkg.Validate()
			if tc.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

func TestRePkgOrRlmPath(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		desc, in    string
		errContains string
	}{
		{
			desc: "Valid p",
			in:   "gno.land/p/path/path",
		},
		{
			desc: "Valid r",
			in:   "gno.land/r/path/path",
		},
		{
			desc: "Valid With Underscore 1",
			in:   "gno.land/r/path/_path",
		},
		{
			desc: "Valid With Underscore 2",
			in:   "gno.land/r/path/pa_th",
		},
		{
			desc: "Valid With Underscore 3",
			in:   "gno.land/r/path/path_",
		},
		{
			desc:        "Invalid x",
			in:          "gno.land/x/path/path",
			errContains: "must be 'p' or 'r'",
		},
		{
			desc:        "No Path 1",
			in:          "gno.land/p",
			errContains: "path must be in the format gno.land/{p|r}/path/...",
		},
		{
			desc:        "No Path 2",
			in:          "gno.land/p/",
			errContains: "path part failed to match",
		},
		{
			desc:        "Invalid domain",
			in:          "github.com/p/path/path",
			errContains: "invalid domain, must be gno.land",
		},
		{
			desc:        "Special Character 1",
			in:          "gno.land/p/p@th/abc/def",
			errContains: "path part failed to match",
		},
		{
			desc:        "Special Character 2",
			in:          "gno.land/p/p&th/abc/def",
			errContains: "path part failed to match",
		},
		{
			desc:        "Special Character 3",
			in:          "gno.land/p/p&%$#h/abc/def",
			errContains: "path part failed to match",
		},
		{
			desc:        "Leading Number",
			in:          "gno.land/p/1Path/abc/def",
			errContains: "path part failed to match",
		},
		{
			desc:        "Uppercase Letters",
			in:          "gno.land/p/PaTh/abc/def",
			errContains: "path part failed to match",
		},
		{
			desc:        "Empty Path Part",
			in:          "gno.land/p/path//def",
			errContains: "path part failed to match",
		},
		{
			desc:        "Trailing Slash",
			in:          "gno.land/p/path/abc/def/",
			errContains: "path part failed to match",
		},
		{
			desc:        "Extra Slash(s)",
			in:          "gno.land/p/path///abc/def",
			errContains: "path part failed to match",
		},
		{
			desc: "Valid Long path",
			in:   "gno.land/r/very/very/very/long/path",
		},
		{
			desc:        "Long Path With Special Character 1",
			in:          "gno.land/r/very/very/very/long/p@th",
			errContains: "path part failed to match",
		},
		{
			desc:        "Long Path With Special Character 2",
			in:          "gno.land/r/very/very/v%ry/long/path",
			errContains: "path part failed to match",
		},
		{
			desc:        "Long Path With Trailing Slash",
			in:          "gno.land/r/very/very/very/long/path/",
			errContains: "path part failed to match",
		},
		{
			desc:        "Long Path With Empty Path Part",
			in:          "gno.land/r/very/very/very//long/path/",
			errContains: "path part failed to match",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := validatePkgOrRlmPath(tc.in)
			if tc.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
