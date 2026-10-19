package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
)

type plainField struct{ A int } // does not implement json.Unmarshaler

func TestDecodeStructFields(t *testing.T) {
	t.Run("non-unmarshaler field: no fallback, no mutation, field name in panic", func(t *testing.T) {
		type params struct {
			Num   int64      `json:"num"`
			Plain plainField `json:"plain"`
		}
		prm := &params{Num: 7, Plain: plainField{A: 5}}
		kvz := []std.KVPair{{Key: []byte("plain"), Value: []byte(`"oops"`)}}

		defer func() {
			r := recover()
			require.NotNil(t, r, "expected panic on undecodable field")
			err, ok := r.(error)
			require.True(t, ok)
			assert.Contains(t, err.Error(), `"plain"`, "panic must name the field")
			// amino failed and plainField has no json.Unmarshaler, so the
			// field must be left untouched (no partial mutation).
			assert.Equal(t, plainField{A: 5}, prm.Plain)
		}()
		decodeStructFields(prm, kvz)
	})

	t.Run("json.Unmarshaler field: string value falls back and decodes", func(t *testing.T) {
		type params struct {
			Num   int64        `json:"num"`
			Price std.GasPrice `json:"price"`
		}
		prm := &params{Num: 7}
		kvz := []std.KVPair{{Key: []byte("price"), Value: []byte(`"1ugnot/1gas"`)}}

		require.NotPanics(t, func() { decodeStructFields(prm, kvz) })
		assert.Equal(t, std.GasPrice{Gas: 1, Price: std.Coin{Denom: "ugnot", Amount: 1}}, prm.Price)
		assert.Equal(t, int64(7), prm.Num)
	})
}

func TestMustParamString(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		got := MustParamString("foo", "bar")
		require.Equal(t, "bar", got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for foo param: expected string, got int",
			func() { MustParamString("foo", 42) },
		)
	})
}

func TestMustParamInt64(t *testing.T) {
	t.Run("valid int64", func(t *testing.T) {
		got := MustParamInt64("num", int64(99))
		require.Equal(t, int64(99), got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for num param: expected int64, got string",
			func() { MustParamInt64("num", "not a number") },
		)
	})
}

func TestMustParamStrings(t *testing.T) {
	t.Run("valid []string", func(t *testing.T) {
		got := MustParamStrings("tags", []string{"a", "b"})
		require.Equal(t, []string{"a", "b"}, got)
	})

	t.Run("wrong type panics", func(t *testing.T) {
		assert.PanicsWithValue(t,
			"invalid type for tags param: expected []string, got string",
			func() { MustParamStrings("tags", "not a slice") },
		)
	})
}
