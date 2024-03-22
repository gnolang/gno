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
		desc, in string
		expected bool
	}{
		{"Valid p", "gno.land/p/path/path", true},
		{"Valid r", "gno.land/r/path/path", true},
		{"Invalid x", "gno.land/x/path/path", false},
		{"Special Character 1", "gno.land/p/p@th/abc/def", false},   // fails
		{"Special Character 2", "gno.land/p/p&th/abc/def", false},   // fails
		{"Special Character 3", "gno.land/p/p&%$#h/abc/def", false}, // fails
		{"Leading Number", "gno.land/p/1Path/abc/def", false},
		{"Uppercase Letters", "gno.land/p/PaTh/abc/def", false},
		{"Empty Path Part", "gno.land/p/path//def", false},     // fails
		{"Trailing Slash", "gno.land/p/path/abc/def/", false},  // fails
		{"Extra Slash(s)", "gno.land/p/path///abc/def", false}, // fails
	} {
		t.Run(tc.desc, func(t *testing.T) {
			res := rePkgOrRlmPath.MatchString(tc.in)
			assert.Equal(t, res, tc.expected)
		})
	}
}
