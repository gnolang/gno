package gnomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func TestRemoveRequireDups(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		in       []*modfile.Require
		expected []*modfile.Require
	}{
		{
			desc: "no_duplicate",
			in: []*modfile.Require{
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
			},
			expected: []*modfile.Require{
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
			},
		},
		{
			desc: "one_duplicate",
			in: []*modfile.Require{
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
			},
			expected: []*modfile.Require{
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
			},
		},
		{
			desc: "multiple_duplicate",
			in: []*modfile.Require{
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.2.0",
					},
				},
			},
			expected: []*modfile.Require{
				{
					Mod: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
				{
					Mod: module.Version{
						Path:    "x.y/w",
						Version: "v1.2.0",
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			in := tc.in
			removeRequireDups(&in)

			assert.Equal(t, tc.expected, in)
		})
	}
}

func TestRemoveReplaceDups(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		in       []*modfile.Replace
		expected []*modfile.Replace
	}{
		{
			desc: "no_duplicate",
			in: []*modfile.Replace{
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
			},
			expected: []*modfile.Replace{
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
				},
			},
		},
		{
			desc: "one_duplicate",
			in: []*modfile.Replace{
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"1"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"2"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"3"},
					},
				},
			},
			expected: []*modfile.Replace{
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"2"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"3"},
					},
				},
			},
		},
		{
			desc: "multiple_duplicate",
			in: []*modfile.Replace{
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"1"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"2"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"3"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"4"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.2.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"5"},
					},
				},
			},
			expected: []*modfile.Replace{
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.0.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"2"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/z",
						Version: "v1.1.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"4"},
					},
				},
				{
					Old: module.Version{
						Path:    "x.y/w",
						Version: "v1.2.0",
					},
					Syntax: &modfile.Line{
						Token: []string{"5"},
					},
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			in := tc.in
			removeReplaceDups(&in)

			assert.Equal(t, tc.expected, in)
		})
	}
}
