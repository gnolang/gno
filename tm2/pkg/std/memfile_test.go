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

	testTable := []struct {
		desc, in string
		expected bool
	}{
		{
			desc:     "Valid p",
			in:       "gno.land/p/path/path",
			expected: true,
		},
		{
			desc:     "Valid r",
			in:       "gno.land/r/path/path",
			expected: true,
		},
		{
			desc:     "Leading Underscore",
			in:       "gno.land/r/path/_path",
			expected: true,
		},
		{
			desc:     "Trailing Underscore",
			in:       "gno.land/r/path/path_",
			expected: true,
		},
		{
			desc:     "Underscore in Between",
			in:       "gno.land/r/path/p_ath",
			expected: true,
		},
		{
			desc:     "Invalid With Underscore 1",
			in:       "gno.land/r/path/_",
			expected: false,
		},
		{
			desc:     "Invalid With Underscore 2",
			in:       "gno.land/r/path/_/_",
			expected: false,
		},
		{
			desc:     "Invalid With Underscore 3",
			in:       "gno.land/r/path/__/path",
			expected: false,
		},
		{
			desc:     "Invalid With Hyphen",
			in:       "gno.land/r/path/pa-th",
			expected: false,
		},
		{
			desc:     "Invalid x",
			in:       "gno.land/x/path/path",
			expected: false,
		},
		{
			desc:     "Missing Path 1",
			in:       "gno.land/p",
			expected: false,
		},
		{
			desc:     "Missing Path 2",
			in:       "gno.land/p/",
			expected: false,
		},
		{
			desc:     "Invalid domain",
			in:       "github.com/p/path/path",
			expected: false,
		},
		{
			desc:     "Special Character 1",
			in:       "gno.land/p/p@th/abc/def",
			expected: false,
		},
		{
			desc:     "Special Character 2",
			in:       "gno.land/p/p&th/abc/def",
			expected: false,
		},
		{
			desc:     "Special Character 3",
			in:       "gno.land/p/p&%$#h/abc/def",
			expected: false,
		},
		{
			desc:     "Leading Number",
			in:       "gno.land/p/1Path/abc/def",
			expected: false,
		},
		{
			desc:     "Uppercase Letters",
			in:       "gno.land/p/PaTh/abc/def",
			expected: false,
		},
		{
			desc:     "Empty Path Part",
			in:       "gno.land/p/path//def",
			expected: false,
		},
		{
			desc:     "Trailing Slash",
			in:       "gno.land/p/path/abc/def/",
			expected: false,
		},
		{
			desc:     "Extra Slash(s)",
			in:       "gno.land/p/path///abc/def",
			expected: false,
		},
		{
			desc:     "Valid Long path",
			in:       "gno.land/r/very/very/very/long/path",
			expected: true,
		},
		{
			desc:     "Long Path With Special Character 1",
			in:       "gno.land/r/very/very/very/long/p@th",
			expected: false,
		},
		{
			desc:     "Long Path With Special Character 2",
			in:       "gno.land/r/very/very/v%ry/long/path",
			expected: false,
		},
		{
			desc:     "Long Path With Trailing Slash",
			in:       "gno.land/r/very/very/very/long/path/",
			expected: false,
		},
		{
			desc:     "Long Path With Empty Path Part",
			in:       "gno.land/r/very/very/very//long/path/",
			expected: false,
		},
	}

	for _, tc := range testTable {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, rePkgOrRlmPath.MatchString(tc.in))
		})
	}
}
