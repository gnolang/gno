package amino

import (
	"reflect"
	"strings"
	"testing"
)

type recSlice struct{ Kids []recSlice }

type recArray struct{ A [1]*recArray }

// recSkipped reaches itself only through fields amino does not decode.
type recSkipped struct{ Mid *recSkipMid }

type recSkipMid struct {
	Back *recSkipped `json:"-"`
	back *recSkipped
}

// recOuter decodes through recRepr, which holds a recOuter again.
type recOuter struct{ X int }

type recRepr struct{ Outer *recOuter }

func (recOuter) MarshalAmino() (recRepr, error) { return recRepr{}, nil }
func (*recOuter) UnmarshalAmino(recRepr) error  { return nil }

// recHolder holds a type that reaches itself but never reaches recHolder.
type recHolder struct{ R recSlice }

// recBare holds itself through its repr alone, so no struct body would count
// its levels.
type recBare struct{ kids []recBare }

func (r recBare) MarshalAmino() ([]recBare, error) { return r.kids, nil }
func (r *recBare) UnmarshalAmino(kids []recBare) error {
	r.kids = kids
	return nil
}

func TestRefuseReprSelfNesting(t *testing.T) {
	defer func() {
		r := recover()
		if s, _ := r.(string); !strings.Contains(s, "with no struct in between") {
			t.Fatalf("recBare: got panic %v, want one refusing its nesting", r)
		}
	}()
	NewCodec().GetTypeInfo(reflect.TypeFor[recBare]())
}

func TestStructInfoRecursive(t *testing.T) {
	cdc := NewCodec()
	for _, tc := range []struct {
		v    any
		want bool
	}{
		{recSlice{}, true},
		{recArray{}, true},
		{recRepr{}, true},
		{recSkipped{}, false},
		{recHolder{}, false},
	} {
		info, err := cdc.GetTypeInfo(reflect.TypeOf(tc.v))
		if err != nil {
			t.Fatal(err)
		}
		if info.Recursive != tc.want {
			t.Errorf("%T: Recursive = %v, want %v", tc.v, info.Recursive, tc.want)
		}
	}
}
