package gnolang

import (
	"reflect"
	"testing"
)

func TestGno2GoValueSliceUsesLength(t *testing.T) {
	const maxcap = 1 << 20
	tests := []struct {
		name string
		tv   *TypedValue
	}{
		{
			name: "byte-backed",
			tv: &TypedValue{
				T: &SliceType{Elt: Uint8Type},
				V: &SliceValue{Base: &ArrayValue{Data: make([]byte, maxcap)}, Maxcap: maxcap},
			},
		},
		{
			name: "list-backed",
			tv: &TypedValue{
				T: &SliceType{Elt: StringType},
				V: &SliceValue{Base: &ArrayValue{List: make([]TypedValue, maxcap)}, Maxcap: maxcap},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := Gno2GoValue(tt.tv, reflect.Value{})
			if rv.Len() != 0 || rv.Cap() != 0 {
				t.Fatalf("converted slice len/cap = %d/%d, want 0/0", rv.Len(), rv.Cap())
			}
		})
	}
}
