package gnoface

import (
	"testing"

	"gno.land/p/demo/ufmt"
)

func TestDraw(t *testing.T) {
	cases := []struct {
		seed     int64
		expected string
	}{
		{
			seed: 42,
			expected: `
  |||||||
 ////////\
 |       |
 | ~   . |
)| X   X |.
 |       |
 |   C   |
 |       |
 |  __/  |
 |       |
 \~~~~~~~/
`[1:],
		},
		{
			seed: 1337,
			expected: `
     s
 /|||||||\
 |       |
 | .   * |
o| ~   ~ |.
 |       |
 |   O   |
 |       |
 |  __/  |
 |       |
 \_______/
`[1:],
		},
		{
			seed: 123456789,
			expected: `
     s
 /~~~~~~~\
 |       |
 | ~   . |
<| ~   ~ |<
 |       |
 |   V   |
 |       |
 |  \_/  |
 |       |
 \-------/
`[1:],
		},
	}
	for _, tc := range cases {
		name := ufmt.Sprintf("%d", tc.seed)
		t.Run(name, func(t *testing.T) {
			got := Draw(tc.seed)
			if got != tc.expected {
				t.Errorf("got %s, expected %s", got, tc.expected)
			}
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
 ////////\
 |       |
 | ~   . |
)| X   X |.
 |       |
 |   C   |
 |       |
 |  __/  |
 |       |
 \~~~~~~~/
` + "```\n",
		},
		{
			path: "1337",
			expected: "Gnoface #1337\n```" + `
     s
 /|||||||\
 |       |
 | .   * |
o| ~   ~ |.
 |       |
 |   O   |
 |       |
 |  __/  |
 |       |
 \_______/
` + "```\n",
		},
		{
			path: "123456789",
			expected: "Gnoface #123456789\n```" + `
     s
 /~~~~~~~\
 |       |
 | ~   . |
<| ~   ~ |<
 |       |
 |   V   |
 |       |
 |  \_/  |
 |       |
 \-------/
` + "```\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := Render(tc.path)
			if got != tc.expected {
				t.Errorf("got %s, expected %s", got, tc.expected)
			}
		})
	}
}
