package amino_test

import (
	"reflect"
	"testing"
	"time"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
)

// These tests lock down the IsByteLengthWrapped predicate against the
// exact kinds of element shapes that genproto2's gen_marshal.go /
// gen_unmarshal.go / gen_size.go used to check inline (the classic
// "Struct || time.Duration || (isListType && non-byte-elem)" pattern —
// see PR #5569 for the class of bug that drift in this predicate
// causes). The predicate is now centralized on *TypeInfo; if any of
// these assertions break, the generator and decoder are no longer
// seeing the same classification.
type ibwLocalStruct struct {
	A int
}

func TestIsByteLengthWrapped_TrueCases(t *testing.T) {
	t.Parallel()
	cdc := amino.NewCodec()

	cases := []struct {
		name string
		rt   reflect.Type
	}{
		{"plain struct", reflect.TypeOf(ibwLocalStruct{})},
		{"time.Time (struct)", reflect.TypeOf(time.Time{})},
		{"time.Duration", reflect.TypeOf(time.Duration(0))},
		{"slice of struct", reflect.TypeOf([]ibwLocalStruct{})},
		{"slice of string", reflect.TypeOf([]string{})},
		{"slice of slice-of-int (nested list)", reflect.TypeOf([][]int{})},
		{"array of struct", reflect.TypeOf([3]ibwLocalStruct{})},
	}
	for _, tc := range cases {
		info, err := cdc.GetTypeInfo(tc.rt)
		if err != nil {
			t.Fatalf("%s: GetTypeInfo: %v", tc.name, err)
		}
		if !info.IsByteLengthWrapped() {
			t.Errorf("%s: expected IsByteLengthWrapped=true, got false", tc.name)
		}
	}
}

func TestIsByteLengthWrapped_FalseCases(t *testing.T) {
	t.Parallel()
	cdc := amino.NewCodec()

	cases := []struct {
		name string
		rt   reflect.Type
	}{
		{"int", reflect.TypeOf(int(0))},
		{"int64", reflect.TypeOf(int64(0))},
		{"uint32", reflect.TypeOf(uint32(0))},
		{"bool", reflect.TypeOf(false)},
		{"string", reflect.TypeOf("")},
		{"float64", reflect.TypeOf(float64(0))},
		{"[]byte", reflect.TypeOf([]byte(nil))},
		{"[4]byte", reflect.TypeOf([4]byte{})},
	}
	for _, tc := range cases {
		info, err := cdc.GetTypeInfo(tc.rt)
		if err != nil {
			t.Fatalf("%s: GetTypeInfo: %v", tc.name, err)
		}
		if info.IsByteLengthWrapped() {
			t.Errorf("%s: expected IsByteLengthWrapped=false, got true", tc.name)
		}
	}
}
