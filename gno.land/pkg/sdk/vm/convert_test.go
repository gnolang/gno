package vm

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

const testMaxAlloc = 1500 * 1000 * 1000

func TestConvertArg2Gno_Primitive(t *testing.T) {
	cases := []struct {
		ValueRep any    // Go  representation
		ArgRep   string // string representation
	}{
		// Boolean
		{true, "true"},
		{false, "false"},

		// int types
		{int(42), "42"},
		{int8(42), "42"},
		{int16(42), "42"},
		{int32(42), "42"},
		{int64(42), "42"},

		// uint types
		{uint(42), "42"},
		{uint8(42), "42"},
		{uint16(42), "42"},
		{uint32(42), "42"},
		{uint64(42), "42"},

		// Float types
		{float32(3.14), "3.14"},
		{float64(3.14), "3.14"},

		// String type
		{"hello world", `hello world`},
	}

	store := setupStore()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.ArgRep, func(t *testing.T) {
			// First, marshal the value and compare it to its string representation
			rv := reflect.ValueOf(tc.ValueRep)
			tvIn := gnolang.Go2GnoValue(store.GetAllocator(), store, rv)
			raw, err := MarshalTypedValueJSON(&tvIn)
			require.NoError(t, err)

			// Unquote intermediary result if necessary to prevent
			// double quoting of a string.
			if isQuotedBytes(raw) {
				raw = unquoteBytes(t, raw)
			}

			// Check if the representation is correct
			require.Equal(t, tc.ArgRep, string(raw))

			// Then, try to use the convert function to get back our Gno value
			// and compare it with the original input
			tvOut, err := convertArgToGno(store, string(raw), tvIn.T)
			require.NoError(t, err)
			require.Equal(t, tvIn.String(), tvOut.String())
		})
	}
}

func TestConvertArg2Gno_Array(t *testing.T) {
	cases := []struct {
		ValueRep any    // Go representation
		ArgRep   string // string representation
	}{
		{[]bool{true, false}, "[true,false]"},
		{[]int{1, 2, 3, 4, 5}, "[1,2,3,4,5]"},
		{[]uint{1, 2, 3, 4, 5}, "[1,2,3,4,5]"},
		{[]float32{1.1, 2.2, 3.3}, "[1.1,2.2,3.3]"},
		{[]string{"hello", "world"}, `["hello","world"]`},

		// XXX: base64 encoded data byte
	}

	store := setupStore()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.ArgRep, func(t *testing.T) {
			// First, marshal the value and compare it to its string representation
			rv := reflect.ValueOf(tc.ValueRep)
			tvIn := gnolang.Go2GnoValue(store.GetAllocator(), store, rv)
			raw, err := MarshalTypedValueJSON(&tvIn)
			require.NoError(t, err)

			// Check if the representation is correct
			require.Equal(t, tc.ArgRep, string(raw))

			// Then, try to use the convert function to get back our Gno value
			// and compare it with the original input
			tvOut, err := convertArgToGno(store, string(raw), tvIn.T)
			require.NoError(t, err)
			require.Equal(t, tvIn.String(), tvOut.String())
		})
	}
}

func TestConvertArg2Gno_Struct(t *testing.T) {
	// Basic struct
	type SimpleStruct struct {
		A bool
		B int
		C string
	}

	// Struct with unexported field
	type UnexportedStruct struct {
		A int
		b string
	}

	// Nested struct
	type NestedStruct struct {
		A int
		B *SimpleStruct
	}

	// Recursive Nested struct
	type RecurseNestedStruct struct {
		A int
		B *RecurseNestedStruct
	}
	recurseNested := &RecurseNestedStruct{A: 42}
	recurseNested.B = recurseNested

	cases := []struct {
		ValueRep any    // Go representation
		ArgRep   string // string representation
	}{
		// Struct with various field values.
		{SimpleStruct{A: true, B: 42, C: "hello"}, `{"A":true,"B":42,"C":"hello"}`},
		{SimpleStruct{A: false, B: 0, C: ""}, `{"A":false,"B":0,"C":""}`},
		{SimpleStruct{A: false, B: 0, C: ""}, `{"A":false,"B":0,"C":""}`},

		// Struct with unexported field
		{UnexportedStruct{A: 42, b: "hidden"}, `{"A":42}`},

		// Struct with nested struct
		{
			NestedStruct{A: 42, B: &SimpleStruct{A: true, B: 43}},
			`{"A":42,"B":{"A":true,"B":43,"C":""}}`,
		},

		// XXX(FIXME): Currently commented out as it causes stack overflow in `Go2GnoValue`
		// Struct with nested and recursive struct
		// {recurseNested, `{A:42}`},
	}

	store := setupStore()
	store.SetStrictGo2GnoMapping(false)
	for _, tc := range cases {
		tc := tc
		t.Run(tc.ArgRep, func(t *testing.T) {
			// First, marshal the value and compare it to its string representation
			rv := reflect.ValueOf(tc.ValueRep)
			tvIn := gnolang.Go2GnoValue(store.GetAllocator(), store, rv)
			raw, err := MarshalTypedValueJSON(&tvIn)
			require.NoError(t, err)

			// Check if the representation is correct
			require.Equal(t, tc.ArgRep, string(raw))

			// Then, use the convert function to get back our Gno value
			// and compare it with the original input
			tvOut, err := convertArgToGno(store, string(raw), tvIn.T)
			require.NoError(t, err)

			// Simple hack to remove the hidden field, as it shouldn't appear in the final result.
			// XXX: It might be better to move this specific test case out to a separate test
			in := strings.ReplaceAll(tvIn.V.String(), "hidden", "")

			// Compare only the value fields here
			require.Equal(t, in, tvOut.V.String())
		})
	}
}

// Prepares the store for testing
func setupStore() gnolang.Store {
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	alloc := gnolang.NewAllocator(testMaxAlloc)
	return gnolang.NewStore(alloc, baseStore, iavlStore)
}

// Checks if the given byte array is enclosed in quotes
func isQuotedBytes(raw []byte) bool {
	return len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"'
}

// Removes the quotes from the given byte array, if present
func unquoteBytes(t *testing.T, raw []byte) []byte {
	t.Helper()

	unquoteRaw, err := strconv.Unquote(string(raw))
	require.NoError(t, err)
	return []byte(unquoteRaw)
}
