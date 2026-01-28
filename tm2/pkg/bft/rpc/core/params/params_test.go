package params

import (
	"encoding/json"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsString(t *testing.T) {
	t.Parallel()

	t.Run("Missing param returns empty string (nil params)", func(t *testing.T) {
		t.Parallel()

		out, err := AsString(nil, 0)
		require.Nil(t, err)

		assert.Equal(t, "", out)
	})

	t.Run("Missing param returns empty string (out of range)", func(t *testing.T) {
		t.Parallel()

		out, err := AsString([]any{"x"}, 10)
		require.Nil(t, err)

		assert.Equal(t, "", out)
	})

	t.Run("String value is returned as-is", func(t *testing.T) {
		t.Parallel()

		out, err := AsString([]any{"hello"}, 0)
		require.Nil(t, err)

		assert.Equal(t, "hello", out)
	})

	t.Run("Fallback path uses JSON then Amino unmarshal", func(t *testing.T) {
		t.Parallel()

		out, err := AsString([]any{[]byte("abc")}, 0)
		require.Nil(t, err)

		assert.Equal(t, "YWJj", out)
	})

	t.Run("Invalid param when JSON marshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsString([]any{func() {}}, 0)
		require.NotNil(t, err)

		assert.Equal(t, "", out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Invalid param when Amino unmarshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsString([]any{map[string]any{"a": 1}}, 0)
		require.NotNil(t, err)

		assert.Equal(t, "", out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})
}

func TestAsBytes(t *testing.T) {
	t.Parallel()

	t.Run("Missing param not required", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes(nil, 0, false)
		require.Nil(t, err)

		assert.Nil(t, out)
	})

	t.Run("Missing param required", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes(nil, 0, true)
		require.NotNil(t, err)

		assert.Nil(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("0x-prefixed hex string is decoded", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes([]any{"0xdeadbeef"}, 0, true)
		require.Nil(t, err)

		assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, out)
	})

	t.Run("Invalid 0x-prefixed hex invalid param", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes([]any{"0xzz"}, 0, true)
		require.NotNil(t, err)

		assert.Nil(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Non-0x string uses Amino semantics for []byte (base64)", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes([]any{"YWJjZA=="}, 0, true)
		require.Nil(t, err)

		assert.Equal(t, []byte("abcd"), out)
	})

	t.Run("Invalid base64 string returns invalid param error", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes([]any{"not-base64"}, 0, true)
		require.NotNil(t, err)

		assert.Nil(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Non-string fallback path uses JSON then Amino unmarshal", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes([]any{[]byte("abcd")}, 0, true)
		require.Nil(t, err)

		assert.Equal(t, []byte("abcd"), out)
	})

	t.Run("Invalid param when Amino unmarshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes([]any{map[string]any{"a": 1}}, 0, true)
		require.NotNil(t, err)

		assert.Nil(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Invalid param when JSON marshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsBytes([]any{make(chan int)}, 0, true)
		require.NotNil(t, err)

		assert.Nil(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})
}

func TestAsInt64(t *testing.T) {
	t.Parallel()

	t.Run("Missing param returns 0 and no error (nil params)", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64(nil, 0)
		require.Nil(t, err)

		assert.Equal(t, int64(0), out)
	})

	t.Run("int64 returns as-is", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{int64(7)}, 0)
		require.Nil(t, err)

		assert.Equal(t, int64(7), out)
	})

	t.Run("int converts to int64", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{int(7)}, 0)
		require.Nil(t, err)

		assert.Equal(t, int64(7), out)
	})

	t.Run("float64 converts to int64 (truncation)", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{float64(7.9)}, 0)
		require.Nil(t, err)

		assert.Equal(t, int64(7), out)
	})

	t.Run("Empty string returns 0 and no error", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{""}, 0)
		require.Nil(t, err)

		assert.Equal(t, int64(0), out)
	})

	t.Run("String parses as base-10 int64", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{"42"}, 0)
		require.Nil(t, err)

		assert.Equal(t, int64(42), out)
	})

	t.Run("Invalid string returns invalid param error", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{"nope"}, 0)
		require.NotNil(t, err)

		assert.Equal(t, int64(0), out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Fallback path uses JSON then Amino unmarshal", func(t *testing.T) {
		t.Parallel()

		type int64String string

		out, err := AsInt64([]any{int64String("123")}, 0)
		require.Nil(t, err)

		assert.Equal(t, int64(123), out)
	})

	t.Run("JSON number and is rejected by legacy Amino", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{json.Number("123")}, 0)
		require.NotNil(t, err)

		assert.Equal(t, int64(0), out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Invalid param when Amino unmarshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{map[string]any{"a": 1}}, 0)
		require.NotNil(t, err)

		assert.Equal(t, int64(0), out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Invalid param when JSON marshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsInt64([]any{make(chan int)}, 0)
		require.NotNil(t, err)

		assert.Equal(t, int64(0), out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})
}

func TestAsBool(t *testing.T) {
	t.Parallel()

	t.Run("Missing param returns false and no error (nil params)", func(t *testing.T) {
		t.Parallel()

		out, err := AsBool(nil, 0)
		require.Nil(t, err)

		assert.False(t, out)
	})

	t.Run("Bool returns as-is", func(t *testing.T) {
		t.Parallel()

		out, err := AsBool([]any{true}, 0)
		require.Nil(t, err)

		assert.True(t, out)
	})

	t.Run("String true/false is accepted (case-insensitive)", func(t *testing.T) {
		t.Parallel()

		out, err := AsBool([]any{"TrUe"}, 0)
		require.Nil(t, err)

		assert.True(t, out)
	})

	t.Run("String false is accepted", func(t *testing.T) {
		t.Parallel()

		out, err := AsBool([]any{"false"}, 0)
		require.Nil(t, err)

		assert.False(t, out)
	})

	t.Run("Invalid string returns invalid param error", func(t *testing.T) {
		t.Parallel()

		out, err := AsBool([]any{"nope"}, 0)
		require.NotNil(t, err)

		assert.False(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Fallback path uses JSON then Amino unmarshal", func(t *testing.T) {
		t.Parallel()

		b := true
		out, err := AsBool([]any{&b}, 0)
		require.Nil(t, err)

		assert.True(t, out)
	})

	t.Run("Invalid param when Amino unmarshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsBool([]any{map[string]any{"a": 1}}, 0)
		require.NotNil(t, err)

		assert.False(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Invalid param when JSON marshal fails", func(t *testing.T) {
		t.Parallel()

		out, err := AsBool([]any{make(chan int)}, 0)
		require.NotNil(t, err)

		assert.False(t, out)
		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})
}
