package gnolang

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertJSONValuePrimitive(t *testing.T) {
	cases := []struct {
		ValueRep string // Go representation
		Expected string // string representation
	}{
		// Boolean - wrapped with protobuf BoolValue
		{"true", `[{"T":"bool","V":{"@type":"/google.protobuf.BoolValue","value":true}}]`},
		{"false", `[{"T":"bool","V":{"@type":"/google.protobuf.BoolValue","value":false}}]`},

		// int types - wrapped with protobuf Int64Value
		{"int(42)", `[{"T":"int","V":{"@type":"/google.protobuf.Int64Value","value":"42"}}]`},
		{"int8(42)", `[{"T":"int8","V":{"@type":"/google.protobuf.Int64Value","value":"42"}}]`},
		{"int16(42)", `[{"T":"int16","V":{"@type":"/google.protobuf.Int64Value","value":"42"}}]`},
		{"int32(42)", `[{"T":"int32","V":{"@type":"/google.protobuf.Int64Value","value":"42"}}]`},
		{"int64(42)", `[{"T":"int64","V":{"@type":"/google.protobuf.Int64Value","value":"42"}}]`},

		// uint types - wrapped with protobuf UInt64Value
		{"uint(42)", `[{"T":"uint","V":{"@type":"/google.protobuf.UInt64Value","value":"42"}}]`},
		{"uint8(42)", `[{"T":"uint8","V":{"@type":"/google.protobuf.UInt64Value","value":"42"}}]`},
		{"uint16(42)", `[{"T":"uint16","V":{"@type":"/google.protobuf.UInt64Value","value":"42"}}]`},
		{"uint32(42)", `[{"T":"uint32","V":{"@type":"/google.protobuf.UInt64Value","value":"42"}}]`},
		{"uint64(42)", `[{"T":"uint64","V":{"@type":"/google.protobuf.UInt64Value","value":"42"}}]`},

		// Float types - converted to string and wrapped with protobuf StringValue
		{"float32(3.14)", `[{"T":"float32","V":{"@type":"/google.protobuf.StringValue","value":"3.14"}}]`},
		{"float64(3.14)", `[{"T":"float64","V":{"@type":"/google.protobuf.StringValue","value":"3.14"}}]`},

		// String type - wrapped with protobuf StringValue
		{`"hello world"`, `[{"T":"string","V":{"@type":"/google.protobuf.StringValue","value":"hello world"}}]`},

		// UntypedRuneType - wrapped with protobuf Int64Value
		{`'A'`, `[{"T":"int32","V":{"@type":"/google.protobuf.Int64Value","value":"65"}}]`},

		// DataByteType (assuming DataByte is an alias for uint8)
		{"uint8(42)", `[{"T":"uint8","V":{"@type":"/google.protobuf.UInt64Value","value":"42"}}]`},
	}

	for _, tc := range cases {
		t.Run(tc.ValueRep, func(t *testing.T) {
			m := NewMachine("testdata", nil)
			defer m.Release()

			nn := m.MustParseFile("testdata.gno",
				fmt.Sprintf(`package testdata; var Value = %s`, tc.ValueRep))
			m.RunFiles(nn)
			m.RunDeclaration(ImportD("testdata", "testdata"))

			tps := m.Eval(Sel(Nx("testdata"), "Value"))
			require.Len(t, tps, 1)

			tv := tps[0]

			rep, err := JSONExportTypedValues([]TypedValue{tv})
			require.NoError(t, err)

			require.Equal(t, tc.Expected, string(rep))
		})
	}
}

func TestConvertJSONValueByteSlice(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata; var Value = []byte("AB")`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	tv := tps[0]
	rep, err := JSONExportTypedValues([]TypedValue{tv})
	require.NoError(t, err)

	// New format expands slices with their base array
	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(rep, &result))
	require.Len(t, result, 1)

	t.Logf("Byte slice output: %s", string(result[0]))
	// Just verify it contains SliceValue structure
	require.Contains(t, string(result[0]), "SliceValue")
}

func TestConvertJSONValueByteArray(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata; var Value = [2]byte{0x41, 0x42}`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	tv := tps[0]
	rep, err := JSONExportTypedValues([]TypedValue{tv})
	require.NoError(t, err)

	// New format expands arrays with Data field
	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(rep, &result))
	require.Len(t, result, 1)

	t.Logf("Byte array output: %s", string(result[0]))
	// Just verify it contains ArrayValue structure
	require.Contains(t, string(result[0]), "ArrayValue")
}

func TestConvertJSONValueStruct(t *testing.T) {
	const StructsFile = `
package testdata

// E struct
type E struct { S string }

func (e *E) String() string { return e.S }
`
	t.Run("null", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno", `package testdata; var Value *E = nil`)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValues([]TypedValue{tv})
		require.NoError(t, err)

		// Verify it's an array with null value
		var result []json.RawMessage
		require.NoError(t, json.Unmarshal(rep, &result))
		require.Len(t, result, 1)
		require.Contains(t, string(result[0]), `"V":null`)
	})

	t.Run("struct value", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValues([]TypedValue{tv})
		require.NoError(t, err)

		// New format expands structs with field names
		var result []json.RawMessage
		require.NoError(t, json.Unmarshal(rep, &result))
		require.Len(t, result, 1)

		t.Logf("Struct value output: %s", string(result[0]))
		// Verify it contains the struct with field name "S"
		require.Contains(t, string(result[0]), `"N":"S"`)
		require.Contains(t, string(result[0]), value)
	})

	t.Run("struct pointer", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = &E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValues([]TypedValue{tv})
		require.NoError(t, err)

		// New format includes PointerValue with expanded base
		var result []json.RawMessage
		require.NoError(t, json.Unmarshal(rep, &result))
		require.Len(t, result, 1)

		t.Logf("Struct pointer output: %s", string(result[0]))
		require.Contains(t, string(result[0]), "PointerValue")
	})
}

func TestConvertJSONValueRecursiveStruct(t *testing.T) {
	const RecursiveValueFile = `
package testdata
type Recursive struct {
        MyString string
	Nested *Recursive
}
var RecursiveStruct = &Recursive{ MyString: "Hello World" }
func init() {
	RecursiveStruct.Nested = RecursiveStruct
}
`
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", RecursiveValueFile)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "RecursiveStruct"))
	require.Len(t, tps, 1)
	tv := tps[0]

	data, err := JSONExportTypedValues([]TypedValue{tv})
	require.NoError(t, err)
	t.Logf("Recursive struct output: %s", string(data))

	// Verify it's valid JSON and handles cycles via RefValue
	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(data, &result))
	require.Len(t, result, 1)

	// Should contain the struct with field names
	require.Contains(t, string(result[0]), `"N":"MyString"`)
	require.Contains(t, string(result[0]), `"N":"Nested"`)
}

// ============================================================================
// JSONObjectInfo OwnerID Tests
// ============================================================================

func TestJSONObjectInfoOwnerID(t *testing.T) {
	// Create valid PkgID for testing
	testPkgID := PkgIDFromPkgPath("gno.land/r/test")

	t.Run("ownerid_shown_when_set", func(t *testing.T) {
		oi := ObjectInfo{
			ID:       ObjectID{PkgID: testPkgID, NewTime: 10},
			OwnerID:  ObjectID{PkgID: testPkgID, NewTime: 5},
			RefCount: 1,
		}

		jsonOI := makeJSONObjectInfo(oi, 0)

		// OwnerID should be set
		require.NotEmpty(t, jsonOI.OwnerID, "OwnerID should be set when non-zero")
		require.Contains(t, jsonOI.OwnerID, ":5", "OwnerID should contain NewTime")
		require.Contains(t, jsonOI.ID, ":10", "ID should contain NewTime")

		// Verify JSON output includes OwnerID
		data, err := json.Marshal(jsonOI)
		require.NoError(t, err)
		require.Contains(t, string(data), `"OwnerID"`)
	})

	t.Run("ownerid_omitted_when_zero", func(t *testing.T) {
		oi := ObjectInfo{
			ID:       ObjectID{PkgID: testPkgID, NewTime: 10},
			OwnerID:  ObjectID{}, // zero
			RefCount: 1,
		}

		jsonOI := makeJSONObjectInfo(oi, 0)

		// OwnerID should be empty
		require.Empty(t, jsonOI.OwnerID, "OwnerID should be empty when zero")

		// Verify JSON output omits OwnerID (due to omitempty)
		data, err := json.Marshal(jsonOI)
		require.NoError(t, err)
		require.NotContains(t, string(data), `"OwnerID"`)
	})

	t.Run("ephemeral_object_with_incremental_id", func(t *testing.T) {
		oi := ObjectInfo{
			ID:       ObjectID{}, // zero - ephemeral
			OwnerID:  ObjectID{}, // zero
			RefCount: 0,
		}

		jsonOI := makeJSONObjectInfo(oi, 5)

		// ID should use incremental format
		require.Equal(t, ":5", jsonOI.ID)
		require.Empty(t, jsonOI.OwnerID)
	})
}

func TestReplaceObjectInfo(t *testing.T) {
	testPkgID := PkgIDFromPkgPath("gno.land/r/test")

	wrapperOI := ObjectInfo{
		ID:       ObjectID{PkgID: testPkgID, NewTime: 100},
		OwnerID:  ObjectID{PkgID: testPkgID, NewTime: 50},
		RefCount: 2,
	}

	t.Run("replace_struct_objectinfo", func(t *testing.T) {
		// Create a JSONStructValue with different ObjectInfo
		jsv := &JSONStructValue{
			ObjectInfo: JSONObjectInfo{ID: ":1"},
			Fields:     []JSONField{{N: "test", V: json.RawMessage(`"value"`)}},
		}

		result := replaceObjectInfo(jsv, wrapperOI)

		replaced, ok := result.(*JSONStructValue)
		require.True(t, ok)
		require.Contains(t, replaced.ObjectInfo.ID, ":100")
		require.Contains(t, replaced.ObjectInfo.OwnerID, ":50")
		require.Equal(t, 2, replaced.ObjectInfo.RefCount)
	})

	t.Run("replace_array_objectinfo", func(t *testing.T) {
		jav := &JSONArrayValue{
			ObjectInfo: JSONObjectInfo{ID: ":2"},
			Elements:   []JSONField{{N: "0", V: json.RawMessage(`1`)}},
		}

		result := replaceObjectInfo(jav, wrapperOI)

		replaced, ok := result.(*JSONArrayValue)
		require.True(t, ok)
		require.Contains(t, replaced.ObjectInfo.ID, ":100")
		require.Contains(t, replaced.ObjectInfo.OwnerID, ":50")
	})

	t.Run("replace_map_objectinfo", func(t *testing.T) {
		jmv := &JSONMapValue{
			ObjectInfo: JSONObjectInfo{ID: ":3"},
			Entries:    []JSONMapEntry{},
		}

		result := replaceObjectInfo(jmv, wrapperOI)

		replaced, ok := result.(*JSONMapValue)
		require.True(t, ok)
		require.Contains(t, replaced.ObjectInfo.ID, ":100")
		require.Contains(t, replaced.ObjectInfo.OwnerID, ":50")
	})
}

func TestExportObjectPreservesWrapperObjectInfo(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	// Create a struct that will be wrapped in HeapItemValue
	code := `package testdata
type Item struct {
	Name string
}
var Value = &Item{Name: "test"}`

	nn := m.MustParseFile("testdata.gno", code)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tvs := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tvs, 1)

	// Get the HeapItemValue
	pv, ok := tvs[0].V.(PointerValue)
	require.True(t, ok, "expected pointer value")
	hiv, ok := pv.Base.(*HeapItemValue)
	require.True(t, ok, "expected heap item value")

	// Set up ObjectInfo on the HeapItemValue to simulate persisted state
	// In real usage, persisted objects have proper ObjectIDs
	hivOI := hiv.GetObjectInfo()

	// Export the HeapItemValue
	opts := JSONExporterOptions{ExportUnexported: true}
	jsonBytes, err := opts.ExportObject(m, hiv)
	require.NoError(t, err)

	// Parse the result
	var result struct {
		ObjectInfo JSONObjectInfo `json:"ObjectInfo"`
	}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	// The exported ObjectInfo should match the HeapItemValue's ObjectInfo,
	// not the inner StructValue's ObjectInfo
	expectedID := makeJSONObjectInfo(*hivOI, 0).ID
	require.Equal(t, expectedID, result.ObjectInfo.ID,
		"exported ObjectInfo.ID should match HeapItemValue's ID")
}
