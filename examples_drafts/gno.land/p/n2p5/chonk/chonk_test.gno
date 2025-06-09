package chonk

import (
	"testing"
)

func TestChonk(t *testing.T) {
	t.Parallel()
	c := New()
	testTable := []struct {
		name   string
		chunks []string
	}{
		{
			name:   "empty",
			chunks: []string{},
		},
		{
			name:   "single chunk",
			chunks: []string{"a"},
		},
		{
			name:   "multiple chunks",
			chunks: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{
			name:   "multiline chunks",
			chunks: []string{"1a\nb\nc\n\n", "d\ne\nf", "g\nh\ni", "j\nk\nl\n\n\n\n"},
		},
		{
			name:   "empty",
			chunks: []string{},
		},
	}
	testChonk := func(t *testing.T, c *Chonk, chunks []string) {
		for _, chunk := range chunks {
			c.Add(chunk)
		}
		scanner := c.Scanner()
		i := 0
		for scanner.Scan() {
			if scanner.Text() != chunks[i] {
				t.Errorf("expected %s, got %s", chunks[i], scanner.Text())
			}
			i++
		}
	}
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			testChonk(t, c, test.chunks)
			c.Flush()
		})
	}
}
