package components

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateFunc_AddSub(t *testing.T) {
	t.Parallel()

	add := funcMap["add"].(func(int, int) int)
	sub := funcMap["sub"].(func(int, int) int)

	assert.Equal(t, 5, add(2, 3))
	assert.Equal(t, 0, add(0, 0))
	assert.Equal(t, -1, add(1, -2))

	assert.Equal(t, 1, sub(3, 2))
	assert.Equal(t, -1, sub(0, 1))
}

func TestTemplateFunc_DerefInt(t *testing.T) {
	t.Parallel()

	derefInt := funcMap["derefInt"].(func(*int) int)

	v := 42
	assert.Equal(t, 42, derefInt(&v))
	assert.Equal(t, 0, derefInt(nil), "nil pointer should return 0")
}

func TestTemplateFunc_QueryHas(t *testing.T) {
	t.Parallel()

	queryHas := funcMap["queryHas"].(func(url.Values, string) bool)

	assert.False(t, queryHas(nil, "any"), "nil values should return false")
	assert.False(t, queryHas(url.Values{}, "missing"))
	assert.True(t, queryHas(url.Values{"state": {""}}, "state"))
}

func TestTemplateFunc_Dict(t *testing.T) {
	t.Parallel()

	dict := funcMap["dict"].(func(...any) (map[string]any, error))

	out, err := dict("a", 1, "b", "two")
	require.NoError(t, err)
	assert.Equal(t, 1, out["a"])
	assert.Equal(t, "two", out["b"])

	_, err = dict("a", 1, "b")
	assert.Error(t, err, "odd number of args should error")

	_, err = dict(1, "value")
	assert.Error(t, err, "non-string key should error")
}

func TestTruncMiddle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		head, tail int
		expected   string
	}{
		{"long string truncated head…tail", "ff61a23bc5d8c018b6c8f29498b1b89435bbeb998", 6, 4, "ff61a2…b998"},
		{"already short string untouched", "abc", 6, 4, "abc"},
		{"exactly threshold untouched", "ff61a23bc", 4, 4, "ff61a23bc"},
		{"tail zero gives head…", "abcdefghij", 3, 0, "abc…"},
		{"head zero gives …tail", "abcdefghij", 0, 3, "…hij"},
		{"empty stays empty", "", 6, 4, ""},
		{"negative bounds clamp to zero", "abcdefghij", -1, -1, "…"},
		{"counts runes, not bytes", "αβγδεζηθικ", 2, 2, "αβ…ικ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, TruncMiddle(tt.input, tt.head, tt.tail))
		})
	}
}
