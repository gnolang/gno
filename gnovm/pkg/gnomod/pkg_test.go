package gnomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortPkgs(t *testing.T) {
	for _, tc := range []struct {
		desc      string
		in        []pkg
		expected  []string
		shouldErr bool
	}{
		{
			desc:     "empty_input",
			in:       []pkg{},
			expected: make([]string, 0),
		}, {
			desc: "no_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{}},
				{name: "pkg3", path: "/path/to/pkg3", requires: []string{}},
			},
			expected: []string{"pkg1", "pkg2", "pkg3"},
		}, {
			desc: "circular_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{"pkg1"}},
			},
			shouldErr: true,
		}, {
			desc: "missing_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
			},
			shouldErr: true,
		}, {
			desc: "valid_dependencies",
			in: []pkg{
				{name: "pkg1", path: "/path/to/pkg1", requires: []string{"pkg2"}},
				{name: "pkg2", path: "/path/to/pkg2", requires: []string{"pkg3"}},
				{name: "pkg3", path: "/path/to/pkg3", requires: []string{}},
			},
			expected: []string{"pkg3", "pkg2", "pkg1"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := SortPkgs(tc.in)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				for i := range tc.expected {
					assert.Equal(t, tc.expected[i], tc.in[i].name)
				}
			}
		})
	}
}
