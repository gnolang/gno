package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortenOID(t *testing.T) {
	t.Parallel()

	const realmA = "ff61a23bc5d8c018b6c8f29498b1b89435bbeb998:11"
	const realmB = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:7"

	tests := []struct {
		name     string
		id, ref  string
		expected string
	}{
		{"same hashlet keeps :N suffix", "ff61a23bc5d8c018b6c8f29498b1b89435bbeb998:1", realmA, ":1"},
		{"different hashlet returns full id", realmB, realmA, realmB},
		{"id without colon untouched", "abcdef", realmA, "abcdef"},
		{"ref without colon returns full id", realmA, "noref", realmA},
		{"empty id returns empty", "", realmA, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, ShortenOID(tt.id, tt.ref))
		})
	}
}

func TestTruncMid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            string
		head, tail       int
		expected         string
	}{
		{"long string truncated head…tail", "ff61a23bc5d8c018b6c8f29498b1b89435bbeb998", 6, 4, "ff61a2…b998"},
		{"already short string untouched", "abc", 6, 4, "abc"},
		{"exactly threshold untouched", "ff61a23bc", 4, 4, "ff61a23bc"},
		{"tail zero gives head…", "abcdefghij", 3, 0, "abc…"},
		{"head zero gives …tail", "abcdefghij", 0, 3, "…hij"},
		{"empty stays empty", "", 6, 4, ""},
		{"negative bounds clamp to zero", "abcdefghij", -1, -1, "…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, truncMid(tt.input, tt.head, tt.tail))
		})
	}
}

func TestTruncOID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            string
		head, tail       int
		expected         string
	}{
		{
			"OID hashlet truncated, :N preserved",
			"ff61a23bc5d8c018b6c8f29498b1b89435bbeb998:11", 6, 4,
			"ff61a2…b998:11",
		},
		{
			"short OID still shows :N",
			"abc:1", 6, 4,
			"abc:1",
		},
		{
			"hash without colon truncated bare",
			"ff61a23bc5d8c018b6c8f29498b1b89435bbeb998", 6, 4,
			"ff61a2…b998",
		},
		{
			"empty returns empty",
			"", 6, 4,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, TruncOID(tt.input, tt.head, tt.tail))
		})
	}
}
