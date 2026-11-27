package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

type plainField struct{ A int } // does not implement json.Unmarshaler

func TestGetStructDecoding(t *testing.T) {
	t.Run("non-unmarshaler field: no fallback, no mutation, field name in panic", func(t *testing.T) {
		type params struct {
			Num   int64      `json:"num"`
			Plain plainField `json:"plain"`
		}
		prm := &params{Num: 7, Plain: plainField{A: 5}}
		env := newTypeInfoTestEnv(t)
		env.keeper.SetString(env.ctx, "ti_a:p:plain", "oops")

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
		env.keeper.GetStruct(env.ctx, "ti_a:p", prm)
	})

	t.Run("json.Unmarshaler field: string value falls back and decodes", func(t *testing.T) {
		type params struct {
			Num   int64        `json:"num"`
			Price std.GasPrice `json:"price"`
		}
		prm := &params{Num: 7}
		env := newTypeInfoTestEnv(t)
		env.keeper.SetString(env.ctx, "ti_a:p:price", "1ugnot/1gas")

		require.NotPanics(t, func() { env.keeper.GetStruct(env.ctx, "ti_a:p", prm) })
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

// Fields amino skips (unexported, or `json:"-"`) make the amino field list
// shorter than the Go field list, so a field's position in tinfo.Fields is
// not its Go struct index. Encode and decode must address fields by
// FieldInfo.Index, or every field after a skipped one lands in the wrong slot.
func TestStructFieldsSkippedFields(t *testing.T) {
	type withUnexported struct {
		hidden int64
		A      int64  `json:"a"`
		B      string `json:"b"`
	}
	type withJSONDash struct {
		Skip int64 `json:"-"`
		A    int64 `json:"a"`
		B    int64 `json:"b"`
	}

	tests := []struct {
		name    string
		in      any
		wantKVs map[string]string
		out     func() any // fresh pointer to decode into
		want    any        // decoded value, skipped fields left zero
	}{
		{
			name:    "unexported leading field",
			in:      withUnexported{hidden: 99, A: 1, B: "x"},
			wantKVs: map[string]string{"a": `"1"`, "b": `"x"`},
			out:     func() any { return &withUnexported{} },
			want:    &withUnexported{A: 1, B: "x"},
		},
		{
			name:    "json dash leading field",
			in:      withJSONDash{Skip: 99, A: 1, B: 2},
			wantKVs: map[string]string{"a": `"1"`, "b": `"2"`},
			out:     func() any { return &withJSONDash{} },
			want:    &withJSONDash{A: 1, B: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kvz := encodeStructFields(tt.in)
			got := map[string]string{}
			for _, kv := range kvz {
				got[string(kv.Key)] = string(kv.Value)
			}
			assert.Equal(t, tt.wantKVs, got, "encoded key/value pairs")

			env := newTypeInfoTestEnv(t)
			env.keeper.SetStruct(env.ctx, "ti_a:p", tt.in)
			out := tt.out()
			env.keeper.GetStruct(env.ctx, "ti_a:p", out)
			assert.Equal(t, tt.want, out, "decoded struct")
		})
	}
}

type paramReadProbe struct {
	store.Store
	keys []string
}

func (p *paramReadProbe) Get(_ *store.GasContext, key []byte) []byte {
	p.keys = append(p.keys, string(key))
	if len(p.keys) == 2 {
		panic("second store read failed")
	}
	return []byte(`"1"`)
}

func TestReadStructFieldsFinishesReadsBeforeDecoding(t *testing.T) {
	out := struct {
		A int64 `json:"a"`
		B int64 `json:"b"`
	}{A: 9, B: 8}
	probe := &paramReadProbe{}

	require.PanicsWithValue(t, "second store read failed", func() {
		readStructFieldsFromStore(nil, &out, probe, "/pv/mod:p:")
	})
	require.Equal(t, []string{"/pv/mod:p:a", "/pv/mod:p:b"}, probe.keys)
	require.Equal(t, int64(9), out.A)
	require.Equal(t, int64(8), out.B)
}
