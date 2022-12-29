package json

import (
	"reflect"
	"testing"

	gno "github.com/gnolang/gno/pkgs/gnolang"
)

func TestMarshal(t *testing.T) {
	var tests = []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "null"},
		{"bool", true, "true"},
		{"bool", false, "false"},
		{"int", int(0), "0"},
		{"int", int8(10), "10"},
		{"int", int16(100), "100"},
		{"int", int32(1000), "1000"},
		{"int", int64(10000), "10000"},
		{"array", [2]int{1, 2}, "[1,2]"},
		{"slice", make([]int, 0), "[]"},
		{"string", "hello", "\"hello\""},
		{"struct", struct {
			A int
			B string
			c string
		}{23, "skidoo", "aa"}, `{"A":23,"B":"skidoo"}`},
		{"struct-tag", struct {
			A int `json:"a"`
			B string
		}{23, "skidoo"}, `{"a":23,"B":"skidoo"}`},
		{"pointer", &struct {
			A int
			B string
		}{23, "skidoo"}, `{"A":23,"B":"skidoo"}`},
		//{"map", map[string]int{"one": 1, "two": 2}, `{"one":1,"two":2}`},
	}
	for _, tt := range tests {
		testMarshal := func(t *testing.T, fn func(v any) (gno.TypedValue, gno.Store)) {
			v, s := fn(tt.in)

			/*
				if v.T != nil {
					fmt.Printf("t: %T v: %T k: %v\n", v.T, v.V, v.T.Kind())
				}
			*/

			got, err := Marshal(v, s)
			//fmt.Printf("got: %s\n", string(got))
			if err != nil {
				t.Errorf("Marshal(%v) error: %v", tt.in, err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal(%v) = %v, want %v", tt.in, string(got), tt.want)
			}
			t.Logf("Marshal(%v) = %v", tt.in, string(got))
		}

		t.Run(tt.name, func(t *testing.T) {
			testMarshal(t, go2GnoTypedValue)
		})
		t.Run(tt.name+"Native", func(t *testing.T) {
			testMarshal(t, go2GnoTypedValueNative)
		})
	}
}

func go2GnoTypedValueNative(v any) (gno.TypedValue, gno.Store) {
	//alloc := gno.NewAllocator(0)
	alloc := (*gno.Allocator)(nil)
	store := gno.NewStore(alloc, nil, nil)
	rv := reflect.ValueOf(v)
	btv := gno.Go2GnoNativeValue(alloc, rv)
	return btv, store
}

func go2GnoTypedValue(v any) (gno.TypedValue, gno.Store) {
	alloc := (*gno.Allocator)(nil)
	//alloc := gno.NewAllocator(0)
	store := gno.NewStore(alloc, nil, nil)
	if v == nil {
		return gno.TypedValue{}, store
	}
	rv := reflect.ValueOf(v)
	btv := gno.Go2GnoValue(alloc, store, rv)
	return btv, store
}
