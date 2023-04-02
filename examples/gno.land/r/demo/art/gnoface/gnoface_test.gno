package gnoface

import (
	"testing"

	"gno.land/p/demo/ufmt"
)

func TestGnoface(t *testing.T) {
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
 \~~~~~~~/`[1:],
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
 \_______/`[1:],
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
 \-------/`[1:],
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
