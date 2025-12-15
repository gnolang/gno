package gnolang

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

func TestConvertJSONValuePrimitive(t *testing.T) {
	cases := []struct {
		ValueRep string // Go representation
		Expected string // string representation
	}{
		// Boolean
		{"true", `{"T":"bool","V":true}`},
		{"false", `{"T":"bool","V":false}`},

		// int types
		{"int(42)", `{"T":"int","V":42}`},
		{"int8(42)", `{"T":"int8","V":42}`},
		{"int16(42)", `{"T":"int16","V":42}`},
		{"int32(42)", `{"T":"int32","V":42}`},
		{"int64(42)", `{"T":"int64","V":42}`},

		// uint types
		{"uint(42)", `{"T":"uint","V":42}`},
		{"uint8(42)", `{"T":"uint8","V":42}`},
		{"uint16(42)", `{"T":"uint16","V":42}`},
		{"uint32(42)", `{"T":"uint32","V":42}`},
		{"uint64(42)", `{"T":"uint64","V":42}`},

		// Float types
		{"float32(3.14)", `{"T":"float32","V":3.14}`},
		{"float64(3.14)", `{"T":"float64","V":3.14}`},

		// String type
		{`"hello world"`, `{"T":"string","V":"hello world"}`},

		// UntypedRuneType
		{`'A'`, `{"T":"int32","V":65}`},

		// DataByteType (assuming DataByte is an alias for uint8)
		{"uint8(42)", `{"T":"uint8","V":42}`},

		// Byte slice - Base is a RefValue reference with ObjectID
		{`[]byte("AB")`, `{"T":"[]uint8","V":{"@type":"/gno.SliceValue","Base":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true},"Offset":"0","Length":"2","Maxcap":"8"}}`},

		// Byte array - exported as RefValue reference
		{`[2]byte{0x41, 0x42}`, `{"T":"[2]uint8","V":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true}}`},
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

			rep, err := JSONExportTypedValue(tv, nil)
			require.NoError(t, err)

			require.Equal(t, tc.Expected, string(rep))
		})
	}
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

		const expected = `{"T":"*RefType{testdata.E}","V":null}`

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno", `package testdata; var Value *E = nil`)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValue(tv, nil)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})

	t.Run("struct value", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"
		// Struct values are exported as RefValue references to break cycles
		const expected = `{"T":"testdata.E","V":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true}}`

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]

		rep, err := JSONExportTypedValue(tv, nil)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
	})

	t.Run("struct pointer", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		const value = "Hello World"
		// Pointer values have their Base as RefValue reference
		const expected = `{"T":"*RefType{testdata.E}","V":{"@type":"/gno.PointerValue","TV":null,"Base":{"@type":"/gno.RefValue","ObjectID":":1","Escaped":true},"Index":"0"}}`

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno",
			fmt.Sprintf(`package testdata; var Value = &E{%q}`, value))
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		tv := tps[0]
		rep, err := JSONExportTypedValue(tv, nil)
		require.NoError(t, err)

		require.Equal(t, string(expected), string(rep))
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

	data, err := JSONExportTypedValue(tv, nil)
	require.NoError(t, err)
	t.Logf("Recursive struct output: %s", string(data))
}

func TestConvertJSONValueRecursiveStructWithSeen(t *testing.T) {
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

	var s struct {
		Refs map[string]json.RawMessage `json:"refs"`
		Val  json.RawMessage            `json:"Value"`
	}

	seen := map[Object]int{}
	data, err := JSONExportTypedValue(tv, seen)
	require.NoError(t, err)
	s.Val = data

	s.Refs = map[string]json.RawMessage{}
	for o := range seen {
		oid := o.GetObjectID()
		if oid.NewTime == 0 {
			continue
		}

		s.Refs[oid.String()] = amino.MustMarshalJSONAny(o)
	}

	data2, err := json.Marshal(s)
	require.NoError(t, err)

	t.Logf("Recursive struct with seen: %s", string(data2))
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
