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

func TestExportValuesMap(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
var Value = map[string]int{"a": 1, "b": 2}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Map output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "MapValue")
}

func TestExportValuesFunc(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
func myFunc(x int) int { return x + 1 }
var Value = myFunc
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Func output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "FuncValue")
}

func TestExportValuesInterface(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
type Stringer interface { String() string }
type MyStr struct { S string }
func (ms MyStr) String() string { return ms.S }
var Value Stringer = MyStr{S: "hello"}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Interface output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "hello")
}

func TestExportValuesIntSlice(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
var Value = []int{10, 20, 30}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Int slice output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "SliceValue")
	require.Contains(t, string(result[0]), "ArrayValue")
}

func TestExportValuesMultiReturn(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
func Multi() (string, int, bool) { return "hi", 99, true }
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Call(Sel(Nx("testdata"), "Multi")))
	require.Len(t, tps, 3)

	bz := exportAndMarshal(t, tps)
	t.Logf("Multi return output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 3)
	require.Contains(t, string(result[0]), "hi")
}

func TestExportValuesListArray(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
var Value = [3]int{1, 2, 3}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("List array output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "ArrayValue")
}

func TestExportObjectStruct(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
type Item struct { Name string; Count int }
var Value = &Item{Name: "widget", Count: 5}
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	// Get the underlying struct object
	pv, ok := tps[0].V.(PointerValue)
	require.True(t, ok)
	obj, ok := pv.Base.(Object)
	require.True(t, ok)

	// ExportObject should expand it inline
	exported := ExportObject(obj)
	require.NotNil(t, exported)

	bz, err := amino.MarshalJSONAny(exported)
	require.NoError(t, err)
	t.Logf("ExportObject output: %s", string(bz))

	require.Contains(t, string(bz), "StructValue")
	require.Contains(t, string(bz), "widget")
}

func TestExportValuesHeapItem(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
func makeHeap() func() int {
	x := 10
	return func() int {
		x++
		return x
	}
}
var Value = makeHeap()
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Heap item output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "FuncValue")
}

func TestExportValuesBoundMethod(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
type Counter struct { N int }
func (c *Counter) Inc() { c.N++ }
var c = &Counter{N: 5}
var Value = c.Inc
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Bound method output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	require.Contains(t, string(result[0]), "BoundMethodValue")
}

func TestExportValuesDeclaredType(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
type MyInt int
func (mi MyInt) Double() MyInt { return mi * 2 }
var Value = MyInt(7)
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	bz := exportAndMarshal(t, tps)
	t.Logf("Declared type output: %s", string(bz))

	var result []json.RawMessage
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Len(t, result, 1)
	// DeclaredType should be exported as RefType
	require.Contains(t, string(result[0]), "RefType")
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

// TestExportCopyTypeWithRefs_InterfaceType exercises the *InterfaceType case
// in exportCopyTypeWithRefs by constructing a raw InterfaceType (not wrapped
// in a DeclaredType) and exporting it.
func TestExportCopyTypeWithRefs_InterfaceType(t *testing.T) {
	seen := make(map[Object]int)
	iface := &InterfaceType{
		PkgPath: "test",
		Methods: []FieldType{
			{Name: "Foo", Type: IntType},
		},
		Generic: "bar",
	}
	copied := exportCopyTypeWithRefs(iface, seen)
	ct, ok := copied.(*InterfaceType)
	require.True(t, ok, "expected *InterfaceType, got %T", copied)
	require.Equal(t, "test", ct.PkgPath)
	require.Equal(t, Name("bar"), ct.Generic)
	require.Len(t, ct.Methods, 1)
	require.Equal(t, Name("Foo"), ct.Methods[0].Name)
}

// TestExportCopyTypeWithRefs_ChanType exercises the *ChanType case
// in exportCopyTypeWithRefs.
func TestExportCopyTypeWithRefs_ChanType(t *testing.T) {
	seen := make(map[Object]int)
	ch := &ChanType{
		Dir: SEND | RECV,
		Elt: IntType,
	}
	copied := exportCopyTypeWithRefs(ch, seen)
	ct, ok := copied.(*ChanType)
	require.True(t, ok, "expected *ChanType, got %T", copied)
	require.Equal(t, SEND|RECV, ct.Dir)
	require.Equal(t, IntType, ct.Elt)
}

// TestExportCopyTypeWithRefs_tupleType exercises the *tupleType case
// in exportCopyTypeWithRefs.
func TestExportCopyTypeWithRefs_tupleType(t *testing.T) {
	seen := make(map[Object]int)
	tt := &tupleType{
		Elts: []Type{
			StringType,
			IntType,
			BoolType,
		},
	}
	copied := exportCopyTypeWithRefs(tt, seen)
	ct, ok := copied.(*tupleType)
	require.True(t, ok, "expected *tupleType, got %T", copied)
	require.Len(t, ct.Elts, 3)
	require.Equal(t, StringType, ct.Elts[0])
	require.Equal(t, IntType, ct.Elts[1])
	require.Equal(t, BoolType, ct.Elts[2])
}

// TestExportCopyValue_BigintValueNilV exercises the BigintValue case in
// exportCopyValue when V is nil (zero-value BigintValue).
func TestExportCopyValue_BigintValueNilV(t *testing.T) {
	seen := make(map[Object]int)
	biv := BigintValue{V: nil}
	copied := exportCopyValue(biv, seen)
	cb, ok := copied.(BigintValue)
	require.True(t, ok, "expected BigintValue, got %T", copied)
	require.Nil(t, cb.V)
}

// TestExportObjectToValue_Block exercises the *Block case in
// exportCopyValue when called via ExportObject. We get a real Block
// by extracting the closure's capture which contains a HeapItemValue
// pointing to a Block through the parent chain.
func TestExportObjectToValue_Block(t *testing.T) {
	m := NewMachine("testdata", nil)
	defer m.Release()

	nn := m.MustParseFile("testdata.gno", `package testdata
func makeClosure() func() int {
	x := 99
	return func() int { return x }
}
var Value = makeClosure()
`)
	m.RunFiles(nn)
	m.RunDeclaration(ImportD("testdata", "testdata"))

	tps := m.Eval(Sel(Nx("testdata"), "Value"))
	require.Len(t, tps, 1)

	// The FuncValue is a closure; walk its captures to find a Block.
	fv, ok := tps[0].V.(*FuncValue)
	require.True(t, ok, "expected *FuncValue, got %T", tps[0].V)

	// Find a *Block in the value tree: check Captures and Parent
	var block *Block
	if fv.Parent != nil {
		if b, ok := fv.Parent.(*Block); ok {
			block = b
		}
	}
	for _, cap := range fv.Captures {
		if hiv, ok := cap.V.(*HeapItemValue); ok {
			if b, ok := hiv.Value.V.(*Block); ok {
				block = b
			}
		}
	}
	// If we still don't have a block, get the package block
	if block == nil {
		// Use the machine's package block instead
		pkg := m.Package
		require.NotNil(t, pkg)
		block = pkg.GetBlock(m.Store)
	}
	require.NotNil(t, block, "could not find a *Block to test")

	exported := ExportObject(block)
	require.NotNil(t, exported)
	eb, ok := exported.(*Block)
	require.True(t, ok, "expected *Block, got %T", exported)
	require.NotEmpty(t, eb.Values)
}

// TestExportObjectToValue_Nil exercises the nil case in exportObjectToValue.
func TestExportObjectToValue_Nil(t *testing.T) {
	exported := ExportObject(nil)
	require.Nil(t, exported)
}
