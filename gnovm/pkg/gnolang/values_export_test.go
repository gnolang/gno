package gnolang

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// helper: export + amino marshal
func exportAndMarshal(t *testing.T, tvs []TypedValue) []byte {
	t.Helper()
	exported := ExportValues(tvs)
	bz, err := amino.MarshalJSON(exported)
	require.NoError(t, err)
	return bz
}

func TestExportValuesPrimitive(t *testing.T) {
	cases := []struct {
		ValueRep string
	}{
		{"true"},
		{"false"},
		{"int(42)"},
		{"int8(42)"},
		{"int16(42)"},
		{"int32(42)"},
		{"int64(42)"},
		{"uint(42)"},
		{"uint8(42)"},
		{"uint16(42)"},
		{"uint32(42)"},
		{"uint64(42)"},
		{`"hello world"`},
		{`'A'`},
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

			bz := exportAndMarshal(t, tps)
			t.Logf("Output: %s", string(bz))

			// Should be valid JSON array with one element
			var result []json.RawMessage
			require.NoError(t, json.Unmarshal(bz, &result))
			require.Len(t, result, 1)
		})
	}
}

func TestExportValuesByteSlice(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata; var Value = []byte("AB")`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Byte slice output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "SliceValue")
}

func TestExportValuesByteArray(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata; var Value = [2]byte{0x41, 0x42}`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Byte array output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "ArrayValue")
}

func TestExportValuesStruct(t *testing.T) {
	const StructsFile = `
package testdata

type E struct { S string }

func (e *E) String() string { return e.S }
`
	t.Run("null pointer", func(t *testing.T) {
		m := NewMachine("testdata", nil)
		defer m.Release()

		nn := m.MustParseFile("struct.gno", StructsFile)
		m.RunFiles(nn)
		nn = m.MustParseFile("testdata.gno", `package testdata; var Value *E = nil`)
		m.RunFiles(nn)
		m.RunDeclaration(ImportD("testdata", "testdata"))

		tps := m.Eval(Sel(Nx("testdata"), "Value"))
		require.Len(t, tps, 1)

		bz := exportAndMarshal(t, tps)
		t.Logf("Null pointer output: %s", string(bz))

		var result []json.RawMessage
		require.NoError(t, json.Unmarshal(bz, &result))
		require.Len(t, result, 1)
		// Amino omits V when nil, so just verify it doesn't contain a value
		require.NotContains(t, string(result[0]), `"V":{"@type":`)
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

		bz := exportAndMarshal(t, tps)
		t.Logf("Struct value output: %s", string(bz))

		var result []json.RawMessage
		require.NoError(t, json.Unmarshal(bz, &result))
		require.Len(t, result, 1)
		// Should contain StructValue with the string value
		require.Contains(t, string(result[0]), "StructValue")
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

		bz := exportAndMarshal(t, tps)
		t.Logf("Struct pointer output: %s", string(bz))

		var result []json.RawMessage
		require.NoError(t, json.Unmarshal(bz, &result))
		require.Len(t, result, 1)
		require.Contains(t, string(result[0]), "PointerValue")
	})
}

func TestExportValuesRecursiveStruct(t *testing.T) {
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

	bz := exportAndMarshal(t, tps)
	t.Logf("Recursive struct output: %s", string(bz))

	// Should be valid JSON and handle cycles via ExportRefValue
	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)

	// Should contain StructValue and ExportRefValue (for the cycle)
	require.Contains(t, string(result[0]), "StructValue")
	require.Contains(t, string(result[0]), "ExportRefValue")
}
