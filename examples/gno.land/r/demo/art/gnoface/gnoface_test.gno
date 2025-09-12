package gnoface

import (
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
)

func TestDraw(t *testing.T) {
	cases := []struct {
		seed     uint64
		expected string
	}{
		{
			seed: 42,
			expected: `
  |||||||
 |||||||||
 |       |
 | .   ~ |
)| v   v |O
 |       |
 |   L   |
 |       |
 |  ___  |
 |       |
 \~~~~~~~/
`[1:],
		},
		{
			seed: 1337,
			expected: `
  .......
 |||||||||
 |       |
 | .   _ |
D| x   X |O
 |       |
 |   ~   |
 |       |
 |  ~~~  |
 |       |
 \~~~~~~~/
`[1:],
		},
		{
			seed: 123456789,
			expected: `
  .......
 ////////\
 |       |
 | ~   * |
|| x   X |o
 |       |
 |   V   |
 |       |
 |   .   |
 |       |
 \-------/
`[1:],
		},
	}
	for _, tc := range cases {
		name := ufmt.Sprintf("%d", tc.seed)
		t.Run(name, func(t *testing.T) {
			got := Draw(tc.seed)
			uassert.Equal(t, string(tc.expected), got)
		})
	}
}

func TestRender(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{
			path: "42",
			expected: "Gnoface #42\n```" + `
  |||||||
 |||||||||
 |       |
 | .   ~ |
)| v   v |O
 |       |
 |   L   |
 |       |
 |  ___  |
 |       |
 \~~~~~~~/
` + "```\n",
		},
		{
			path: "1337",
			expected: "Gnoface #1337\n```" + `
  .......
 |||||||||
 |       |
 | .   _ |
D| x   X |O
 |       |
 |   ~   |
 |       |
 |  ~~~  |
 |       |
 \~~~~~~~/
` + "```\n",
		},
		{
			path: "123456789",
			expected: "Gnoface #123456789\n```" + `
  .......
 ////////\
 |       |
 | ~   * |
|| x   X |o
 |       |
 |   V   |
 |       |
 |   .   |
 |       |
 \-------/
` + "```\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := Render(tc.path)
			uassert.Equal(t, tc.expected, got)
		})
	}
}
