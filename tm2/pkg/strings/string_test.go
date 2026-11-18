package strings

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestStringInSlice(t *testing.T) {
	t.Parallel()

	assert.True(t, StringInSlice("a", []string{"a", "b", "c"}))
	assert.False(t, StringInSlice("d", []string{"a", "b", "c"}))
	assert.True(t, StringInSlice("", []string{""}))
	assert.False(t, StringInSlice("", []string{}))
}

func TestIsASCIIText(t *testing.T) {
	t.Parallel()

	notASCIIText := []string{
		"", "\xC2", "\xC2\xA2", "\xFF", "\x80", "\xF0", "\n", "\t",
	}
	for _, v := range notASCIIText {
		assert.False(t, IsASCIIText(v), "%q is not ascii-text", v)
	}
	asciiText := []string{
		" ", ".", "x", "$", "_", "abcdefg;", "-", "0x00", "0", "123",
	}
	for _, v := range asciiText {
		assert.True(t, IsASCIIText(v), "%q is ascii-text", v)
	}
}

func TestASCIITrim(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ASCIITrim(" "), "")
	assert.Equal(t, ASCIITrim(" a"), "a")
	assert.Equal(t, ASCIITrim("a "), "a")
	assert.Equal(t, ASCIITrim(" a "), "a")
	assert.Panics(t, func() { ASCIITrim("\xC2\xA2") })
}

func TestTrimN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"shorter than n", "hello", 10, "hello"},
		{"equal to n", "hello", 5, "hello"},
		{"longer than n", "hello world", 8, "hello..."},
		{"n equals three", "hello", 3, "..."},
		{"n is four", "hello", 4, "h..."},
		{"n two no room for ellipsis", "hello", 2, "he"},
		{"n one no room for ellipsis", "hello", 1, "h"},
		{"n zero", "hello", 0, ""},
		{"n negative", "hello", -1, ""},
		{"empty string", "", 5, ""},
		{"empty string n zero", "", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TrimN(tt.s, tt.n)
			require.Equal(t, tt.want, got)
			require.LessOrEqual(t, len(got), max(tt.n, 0),
				"result must fit within n bytes")
		})
	}
}

func TestStringSliceEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a    []string
		b    []string
		want bool
	}{
		{[]string{"hello", "world"}, []string{"hello", "world"}, true},
		{[]string{"test"}, []string{"test"}, true},
		{[]string{"test1"}, []string{"test2"}, false},
		{[]string{"hello", "world."}, []string{"hello", "world!"}, false},
		{[]string{"only 1 word"}, []string{"two", "words!"}, false},
		{[]string{"two", "words!"}, []string{"only 1 word"}, false},
	}
	for i, tt := range tests {
		require.Equal(t, tt.want, StringSliceEqual(tt.a, tt.b),
			"StringSliceEqual failed on test %d", i)
	}
}
