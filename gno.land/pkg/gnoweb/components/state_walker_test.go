package components

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// qpkgFixture mirrors the test fixture from misc/gnojs/src/decode.test.ts —
// real Go test output of vm/qpkg_json. Names match what the TS decoder asserts
// on, so any divergence between the Go and TS decoders surfaces here.
const qpkgFixture = `{
  "names": ["MyStruct", "myInt", "myStr", "myStruct", "init.4", "Render"],
  "values": [
    {"T": {"@type": "/gno.TypeType"}, "V": {"@type": "/gno.TypeValue", "Type": {"@type": "/gno.DeclaredType", "PkgPath": "gno.land/r/test/qpkg", "Name": "MyStruct", "Base": {"@type": "/gno.StructType", "PkgPath": "gno.land/r/test/qpkg", "Fields": [{"Name": "Name", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""}, {"Name": "Age", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}]}, "Methods": []}}},
    {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "KgAAAAAAAAA="},
    {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "hello"}},
    {"T": {"@type": "/gno.RefType", "ID": "gno.land/r/test/qpkg.MyStruct"}, "V": {"@type": "/gno.RefValue", "ObjectID": "715383ba05505afed61caa873216e2ee896bede9:10"}},
    {"T": {"@type": "/gno.FuncType", "Params": [], "Results": []}, "V": {"@type": "/gno.RefValue", "ObjectID": "715383ba05505afed61caa873216e2ee896bede9:7"}},
    {"T": {"@type": "/gno.FuncType", "Params": [{"Name": "path", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""}], "Results": [{"Name": ".res.0", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""}]}, "V": {"@type": "/gno.RefValue", "ObjectID": "715383ba05505afed61caa873216e2ee896bede9:9"}}
  ]
}`

func TestDecodePkgJSON_TopLevel(t *testing.T) {
	t.Parallel()

	nodes, err := DecodePkgJSON([]byte(qpkgFixture))
	require.NoError(t, err)
	require.Len(t, nodes, 6, "should decode 6 top-level nodes")

	// MyStruct — TypeValue
	assert.Equal(t, "MyStruct", nodes[0].Name)
	assert.Equal(t, "type", nodes[0].Kind, "TypeValue kind")
	assert.NotEmpty(t, nodes[0].Value, "TypeValue should display its type")

	// myInt — int = 42 (N base64 = 42 LE int64)
	assert.Equal(t, "myInt", nodes[1].Name)
	assert.Equal(t, "int", nodes[1].Type)
	assert.Equal(t, "primitive", nodes[1].Kind)
	assert.Equal(t, "42", nodes[1].Value)
	assert.False(t, nodes[1].Expandable)

	// myStr — string = "hello"
	assert.Equal(t, "myStr", nodes[2].Name)
	assert.Equal(t, "string", nodes[2].Type)
	assert.Equal(t, "primitive", nodes[2].Kind)
	assert.Equal(t, `"hello"`, nodes[2].Value)

	// myStruct — RefType + RefValue → expandable, links to a separate page
	assert.Equal(t, "myStruct", nodes[3].Name)
	assert.True(t, nodes[3].Expandable)
	assert.Equal(t, "715383ba05505afed61caa873216e2ee896bede9:10", nodes[3].ObjectID)
	assert.Equal(t, "gno.land/r/test/qpkg.MyStruct", nodes[3].TypeID)

	// init.4 — FuncType with RefValue
	assert.Equal(t, "init.4", nodes[4].Name)
	assert.Equal(t, "func", nodes[4].Kind)

	// Render — FuncType(path string) string
	assert.Equal(t, "Render", nodes[5].Name)
	assert.Equal(t, "func", nodes[5].Kind)
	assert.Contains(t, nodes[5].Type, "func", "should look like a func signature")
	assert.Contains(t, nodes[5].Type, "string", "signature should mention string")
}

// TestDecodeObjectJSON_Struct exercises the qobject_json shape: the response
// wraps a single Value (here a StructValue), and decoding produces the
// children of that value as a flat list with positional names.
func TestDecodeObjectJSON_Struct(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:8",
		"value": {
			"@type": "/gno.StructValue",
			"ObjectInfo": {"ID": "ffffffffffffffffffffffffffffffffffffffff:8"},
			"Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="},
				{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "test"}}
			]
		}
	}`

	nodes, err := DecodeObjectJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 2)

	// Field 0: int = 1
	assert.Equal(t, "0", nodes[0].Name)
	assert.Equal(t, "int", nodes[0].Type)
	assert.Equal(t, "1", nodes[0].Value)

	// Field 1: string = "test"
	assert.Equal(t, "1", nodes[1].Name)
	assert.Equal(t, "string", nodes[1].Type)
	assert.Equal(t, `"test"`, nodes[1].Value)
}

// TestDecodeCycleRef ensures ExportRefValue produces a non-expandable
// "<cycle :N>" placeholder while preserving the type's kind (pointer here),
// mirroring TS testDecodeCycleRef behavior.
func TestDecodeCycleRef(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["self"],
		"values": [
			{"T": {"@type": "/gno.PointerType", "Elt": {"@type": "/gno.RefType", "ID": "gno.land/r/test.Node"}},
			 "V": {"@type": "/gno.ExportRefValue", "ObjectID": ":1"}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	assert.Equal(t, "self", nodes[0].Name)
	assert.Equal(t, "pointer", nodes[0].Kind, "kind preserved from PointerType")
	assert.False(t, nodes[0].Expandable)
	assert.Equal(t, "<cycle :1>", nodes[0].Value)
}

// TestDecodeHeapItemUnwrap verifies that a HeapItemValue is transparently
// unwrapped — its child becomes the visible node.
func TestDecodeHeapItemUnwrap(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["wrapped"],
		"values": [
			{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "V": {"@type": "/gno.HeapItemValue", "Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "BwAAAAAAAAA="}}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	// Unwrapped: should look like a plain int = 7, no "HeapItem" leakage.
	assert.Equal(t, "wrapped", nodes[0].Name)
	assert.Equal(t, "primitive", nodes[0].Kind)
	assert.Equal(t, "7", nodes[0].Value)
}

// TestDecodeNilT covers the case where T is missing — the node should render
// as a nil placeholder instead of crashing.
func TestDecodeNilT(t *testing.T) {
	t.Parallel()

	const fixture = `{"names": ["empty"], "values": [{}]}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	assert.Equal(t, "empty", nodes[0].Name)
	assert.Equal(t, "nil", nodes[0].Kind)
	assert.False(t, nodes[0].Expandable)
}

// TestDecodeMap exercises an inline MapValue: each entry becomes a child whose
// name is the rendered key and whose value is the rendered value.
func TestDecodeMap(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["m"],
		"values": [
			{"T": {"@type": "/gno.MapType", "Key": {"@type": "/gno.PrimitiveType", "value": "16"}, "Value": {"@type": "/gno.PrimitiveType", "value": "32"}},
			 "V": {"@type": "/gno.MapValue", "List": {"List": [
				{"Key": {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "alice"}},
				 "Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "HgAAAAAAAAA="}},
				{"Key": {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "bob"}},
				 "Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "GQAAAAAAAAA="}}
			 ]}}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	m := nodes[0]
	assert.Equal(t, "m", m.Name)
	assert.Equal(t, "map", m.Kind)
	require.NotNil(t, m.Length)
	assert.Equal(t, 2, *m.Length)
	require.Len(t, m.Children, 2)

	assert.Equal(t, `"alice"`, m.Children[0].Name)
	assert.Equal(t, "30", m.Children[0].Value)
	assert.Equal(t, `"bob"`, m.Children[1].Name)
	assert.Equal(t, "25", m.Children[1].Value)
}

// TestDecodeSliceInline covers a SliceValue whose Base is an inline ArrayValue
// (not a RefValue) — children should be drawn from the array list, sliced by
// Offset/Length.
func TestDecodeSliceInline(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["nums"],
		"values": [
			{"T": {"@type": "/gno.SliceType", "Elt": {"@type": "/gno.PrimitiveType", "value": "32"}, "Vrd": false},
			 "V": {"@type": "/gno.SliceValue",
				"Base": {"@type": "/gno.ArrayValue", "List": [
					{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="},
					{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AgAAAAAAAAA="},
					{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AwAAAAAAAAA="}
				]},
				"Offset": "0", "Length": "3", "Maxcap": "3"}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	s := nodes[0]
	assert.Equal(t, "slice", s.Kind)
	require.NotNil(t, s.Length)
	assert.Equal(t, 3, *s.Length)
	require.Len(t, s.Children, 3)
	assert.Equal(t, "1", s.Children[0].Value)
	assert.Equal(t, "2", s.Children[1].Value)
	assert.Equal(t, "3", s.Children[2].Value)
}

// TestDecodeStringTruncated checks the long-string truncation: strings longer
// than 256 chars get truncated with an ellipsis. Mirrors decode.ts behavior.
func TestDecodeStringTruncated(t *testing.T) {
	t.Parallel()

	long := makeLongString(300)
	fixture := `{"names": ["s"], "values": [{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "` + long + `"}}]}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	assert.Less(t, len(nodes[0].Value), len(long)+10, "long strings should be truncated, not echoed")
	assert.Contains(t, nodes[0].Value, "...", "truncated strings end with ellipsis")
}

func makeLongString(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = 'a'
	}
	return string(out)
}

// TestDecodeInlineStruct mirrors TS testDecodeInlineStruct: an inline
// StructValue + StructType produces children with the field names from the
// type declaration.
func TestDecodeInlineStruct(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["s"],
		"values": [
			{"T": {"@type": "/gno.StructType", "PkgPath": "gno.land/r/test", "Fields": [
				{"Name": "X", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""},
				{"Name": "Y", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""}
			]},
			 "V": {"@type": "/gno.StructValue", "Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "CQAAAAAAAAA="},
				{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "hi"}}
			 ]}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	s := nodes[0]
	assert.Equal(t, "struct", s.Kind)
	require.NotNil(t, s.Length)
	assert.Equal(t, 2, *s.Length)
	require.Len(t, s.Children, 2)

	assert.Equal(t, "X", s.Children[0].Name, "field name X from StructType")
	assert.Equal(t, "9", s.Children[0].Value, "X = 9")
	assert.Equal(t, "Y", s.Children[1].Name)
	assert.Equal(t, `"hi"`, s.Children[1].Value)
}

// TestDecodePointerRefValue mirrors TS testDecodePointerRefValue: a pointer
// whose Base is a RefValue is expandable and exposes the target ObjectID.
func TestDecodePointerRefValue(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["ptr"],
		"values": [
			{"T": {"@type": "/gno.PointerType", "Elt": {"@type": "/gno.RefType", "ID": "gno.land/r/test.Foo"}},
			 "V": {"@type": "/gno.PointerValue", "TV": null,
				"Base": {"@type": "/gno.RefValue", "ObjectID": "ffffffffffffffffffffffffffffffffffffffff:5"},
				"Index": "0"}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	p := nodes[0]
	assert.Equal(t, "pointer", p.Kind)
	assert.True(t, p.Expandable)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:5", p.ObjectID)
}

// TestDecodeSliceRefBase mirrors TS testDecodeSliceRefBase: a slice whose
// Base is a RefValue (stored array) becomes expandable with ObjectID set.
func TestDecodeSliceRefBase(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["items"],
		"values": [
			{"T": {"@type": "/gno.SliceType", "Elt": {"@type": "/gno.PrimitiveType", "value": "32"}, "Vrd": false},
			 "V": {"@type": "/gno.SliceValue",
				"Base": {"@type": "/gno.RefValue", "ObjectID": "ffffffffffffffffffffffffffffffffffffffff:3"},
				"Offset": "0", "Length": "5", "Maxcap": "8"}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	s := nodes[0]
	assert.Equal(t, "slice", s.Kind)
	require.NotNil(t, s.Length)
	assert.Equal(t, 5, *s.Length)
	assert.True(t, s.Expandable)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:3", s.ObjectID)
}

// TestDecodeFuncInline mirrors TS testDecodeFuncInline: an inline FuncValue
// (with Source RefNode) reports its file/line span.
func TestDecodeFuncInline(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["myFunc"],
		"values": [
			{"T": {"@type": "/gno.FuncType",
				"Params": [{"Name": "x", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}],
				"Results": [{"Name": ".res.0", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}]},
			 "V": {"@type": "/gno.FuncValue",
				"Type": {"@type": "/gno.FuncType",
					"Params": [{"Name": "x", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}],
					"Results": [{"Name": ".res.0", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}]},
				"Name": "myFunc",
				"Source": {"@type": "/gno.RefNode",
					"Location": {"PkgPath": "gno.land/r/test", "File": "test.gno",
						"Span": {"Pos": {"Line": "5", "Column": "1"}, "End": {"Line": "7", "Column": "1"}, "Num": "0"}}}}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	f := nodes[0]
	assert.Equal(t, "func", f.Kind)
	require.NotNil(t, f.Source, "inline func should expose a source location")
	assert.Equal(t, "test.gno", f.Source.File)
	assert.Equal(t, 5, f.Source.StartLine)
	assert.Equal(t, 7, f.Source.EndLine)
}

// TestDecodeFuncRefValue mirrors TS testDecodeFuncRefValue: a FuncType whose
// V is a RefValue is expandable (its body is fetched on a separate page) and
// the type display includes a func signature.
func TestDecodeFuncRefValue(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["Render"],
		"values": [
			{"T": {"@type": "/gno.FuncType",
				"Params": [{"Name": "path", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""}],
				"Results": [{"Name": ".res.0", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""}]},
			 "V": {"@type": "/gno.RefValue", "ObjectID": "ffffffffffffffffffffffffffffffffffffffff:9"}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	f := nodes[0]
	assert.Equal(t, "func", f.Kind)
	assert.True(t, f.Expandable)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:9", f.ObjectID)
	assert.Contains(t, f.Type, "func(", "type display should include signature")
}

// TestDecodeClosureWithCaptures mirrors TS testDecodeClosureWithCaptures:
// FuncValue with non-empty Captures becomes kind="closure", with one child
// per capture.
func TestDecodeClosureWithCaptures(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["stepper"],
		"values": [
			{"T": {"@type": "/gno.FuncType", "Params": [],
				"Results": [{"Name": ".res.0", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}]},
			 "V": {"@type": "/gno.FuncValue",
				"Type": {"@type": "/gno.FuncType", "Params": [],
					"Results": [{"Name": ".res.0", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}]},
				"Name": "",
				"Captures": [
					{"T": {"@type": "/gno.heapItemType"}, "V": {"@type": "/gno.RefValue", "ObjectID": "ffffffffffffffffffffffffffffffffffffffff:13"}}
				],
				"Source": {"@type": "/gno.RefNode",
					"Location": {"PkgPath": "gno.land/r/test", "File": "test.gno",
						"Span": {"Pos": {"Line": "17", "Column": "12"}, "End": {"Line": "20", "Column": "3"}, "Num": "0"}}}}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	c := nodes[0]
	assert.Equal(t, "closure", c.Kind)
	assert.True(t, c.Expandable)
	require.NotNil(t, c.Source, "closure should have source")
	require.Len(t, c.Children, 1, "one capture")
	assert.Equal(t, "value", c.Children[0].Name)
	assert.True(t, c.Children[0].Expandable, "RefValue capture is expandable")
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:13", c.Children[0].ObjectID)
}

// TestDecodeFuncNoCapturesNotClosure mirrors TS test of same name: a FuncValue
// with empty Captures is "func", not "closure".
func TestDecodeFuncNoCapturesNotClosure(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["init"],
		"values": [
			{"T": {"@type": "/gno.FuncType", "Params": [], "Results": []},
			 "V": {"@type": "/gno.FuncValue",
				"Type": {"@type": "/gno.FuncType", "Params": [], "Results": []},
				"Name": "init",
				"Captures": [],
				"Source": {"@type": "/gno.RefNode",
					"Location": {"PkgPath": "gno.land/r/test", "File": "test.gno",
						"Span": {"Pos": {"Line": "1", "Column": "1"}, "End": {"Line": "3", "Column": "1"}, "Num": "0"}}}}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	f := nodes[0]
	assert.Equal(t, "func", f.Kind, "no captures -> not a closure")
	assert.Empty(t, f.Children)
}

// TestDecodeClosureMultipleCaptures: a closure with several captures gets
// one child per capture, in order.
func TestDecodeClosureMultipleCaptures(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["accumulator"],
		"values": [
			{"T": {"@type": "/gno.FuncType",
				"Params": [{"Name": "val", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}],
				"Results": []},
			 "V": {"@type": "/gno.FuncValue",
				"Type": {"@type": "/gno.FuncType",
					"Params": [{"Name": "val", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}],
					"Results": []},
				"Name": "",
				"Captures": [
					{"T": {"@type": "/gno.heapItemType"}, "V": {"@type": "/gno.RefValue", "ObjectID": "ffffffffffffffffffffffffffffffffffffffff:16"}},
					{"T": {"@type": "/gno.heapItemType"}, "V": {"@type": "/gno.RefValue", "ObjectID": "ffffffffffffffffffffffffffffffffffffffff:17"}}
				],
				"Source": {"@type": "/gno.RefNode",
					"Location": {"PkgPath": "gno.land/r/test", "File": "test.gno",
						"Span": {"Pos": {"Line": "23", "Column": "16"}, "End": {"Line": "25", "Column": "3"}, "Num": "0"}}}}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	c := nodes[0]
	assert.Equal(t, "closure", c.Kind)
	require.Len(t, c.Children, 2)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:16", c.Children[0].ObjectID)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:17", c.Children[1].ObjectID)
}

// TestDecodeArrayBytes covers byte arrays (Data field) — they should render
// as a "[N]byte{...}" placeholder, not enumerate every byte.
func TestDecodeArrayBytes(t *testing.T) {
	t.Parallel()

	// "hello" = base64 "aGVsbG8=" → 5 bytes.
	const fixture = `{
		"names": ["data"],
		"values": [
			{"T": {"@type": "/gno.ArrayType", "Len": "5", "Elt": {"@type": "/gno.PrimitiveType", "value": "4096"}, "Vrd": false},
			 "V": {"@type": "/gno.ArrayValue", "Data": "aGVsbG8="}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	a := nodes[0]
	assert.Equal(t, "array", a.Kind)
	require.NotNil(t, a.Length)
	assert.Equal(t, 5, *a.Length)
	assert.Equal(t, "[5]byte{...}", a.Value, "byte arrays summarized, not expanded")
	assert.False(t, a.Expandable)
}

// TestDecodePrimitives_Variants exercises every primitive type kind that
// Gno realms can hold. Each variant has its own N-field decoding path
// (different sign, width, float bit-pattern) — silent regressions per type
// would mean wrong values displayed in production for things like token
// balances (uint64) or boolean flags. Table-driven to keep coverage tight.
func TestDecodePrimitives_Variants(t *testing.T) {
	t.Parallel()

	// PrimitiveType numeric values from gnovm/pkg/gnolang/types.go.
	const (
		pBool    = "4"
		pInt     = "32"
		pInt8    = "64"
		pInt16   = "128"
		pInt32   = "512"
		pInt64   = "1024"
		pUint    = "2048"
		pUint8   = "4096"
		pUint16  = "16384"
		pUint32  = "32768"
		pUint64  = "65536"
		pFloat32 = "131072"
		pFloat64 = "262144"
	)

	// Build N as 8-byte little-endian base64.
	n8 := func(v uint64) string {
		buf := make([]byte, 8)
		for i := 0; i < 8; i++ {
			buf[i] = byte(v >> (8 * i))
		}
		return base64.StdEncoding.EncodeToString(buf)
	}

	cases := []struct {
		name      string
		typeValue string
		nField    string
		wantType  string
		wantValue string
	}{
		{"bool true", pBool, n8(1), "bool", "true"},
		{"bool false", pBool, n8(0), "bool", "false"},
		{"int positive", pInt, n8(42), "int", "42"},
		{"int8 -1", pInt8, n8(0xFF), "int8", "-1"},
		{"int16 -1", pInt16, n8(0xFFFF), "int16", "-1"},
		{"int32 -1", pInt32, n8(0xFFFFFFFF), "int32", "-1"},
		{"int64 max-ish", pInt64, n8(0x7FFFFFFFFFFFFFFF), "int64", "9223372036854775807"},
		{"uint", pUint, n8(100), "uint", "100"},
		{"uint8 max", pUint8, n8(255), "uint8", "255"},
		{"uint16", pUint16, n8(65535), "uint16", "65535"},
		{"uint32", pUint32, n8(4000000000), "uint32", "4000000000"},
		{"uint64 huge", pUint64, n8(18446744073709551615), "uint64", "18446744073709551615"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fixture := fmt.Sprintf(
				`{"names":["x"],"values":[{"T":{"@type":"/gno.PrimitiveType","value":"%s"},"N":"%s"}]}`,
				c.typeValue, c.nField,
			)
			nodes, err := DecodePkgJSON([]byte(fixture))
			require.NoError(t, err)
			require.Len(t, nodes, 1)
			assert.Equal(t, c.wantType, nodes[0].Type, "type display")
			assert.Equal(t, "primitive", nodes[0].Kind)
			assert.Equal(t, c.wantValue, nodes[0].Value, "decoded primitive value")
		})
	}
}

// TestDecodeArrayInline covers a non-byte array with explicit List entries —
// e.g. `[3]string{...}` rendered with one child per element. Distinct from
// the Data path tested in TestDecodeArrayBytes.
func TestDecodeArrayInline(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["arr"],
		"values": [
			{"T": {"@type": "/gno.ArrayType", "Len": "3", "Elt": {"@type": "/gno.PrimitiveType", "value": "16"}, "Vrd": false},
			 "V": {"@type": "/gno.ArrayValue", "List": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "a"}},
				{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "b"}},
				{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "c"}}
			 ]}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	a := nodes[0]
	assert.Equal(t, "array", a.Kind)
	require.NotNil(t, a.Length)
	assert.Equal(t, 3, *a.Length)
	require.Len(t, a.Children, 3)
	assert.Equal(t, "0", a.Children[0].Name, "positional index name")
	assert.Equal(t, `"a"`, a.Children[0].Value)
	assert.Equal(t, `"c"`, a.Children[2].Value)
}

// TestDecodePointerInline covers `*int = &x` style pointers — TV is set
// (inline target), no RefValue base. The TS test only covered the
// RefValue-base case; this is the sibling code path.
func TestDecodePointerInline(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"names": ["p"],
		"values": [
			{"T": {"@type": "/gno.PointerType", "Elt": {"@type": "/gno.PrimitiveType", "value": "32"}},
			 "V": {"@type": "/gno.PointerValue",
				"TV": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "BwAAAAAAAAA="},
				"Index": "0"}}
		]
	}`

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	p := nodes[0]
	assert.Equal(t, "pointer", p.Kind)
	require.Len(t, p.Children, 1, "inline pointer exposes its target as child")
	assert.Equal(t, "*", p.Children[0].Name)
	assert.Equal(t, "7", p.Children[0].Value)
}

// TestDecodeObjectJSON_Map covers the qobject_json path for a MapValue —
// production hits this when the user navigates to a stored map's page.
// Distinct from the StructValue path already tested.
func TestDecodeObjectJSON_Map(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:8",
		"value": {
			"@type": "/gno.MapValue",
			"ObjectInfo": {"ID": "ffffffffffffffffffffffffffffffffffffffff:8"},
			"List": {"List": [
				{"Key": {"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "k"}},
				 "Value": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="}}
			]}
		}
	}`

	nodes, err := DecodeObjectJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	assert.Equal(t, `"k"`, nodes[0].Name, "map key as child name")
	assert.Equal(t, "1", nodes[0].Value)
}

// TestDecodeObjectJSONWithType_LabelsStructFields exercises the field-name
// resolution path: when a qtype_json response is provided alongside the
// object JSON, struct fields render with their declared names instead of
// "0", "1", "2" — Amino strips named-type info during ExportValues, so
// we recover it via the separate type fetch.
func TestDecodeObjectJSONWithType_LabelsStructFields(t *testing.T) {
	t.Parallel()

	// Object JSON: a HeapItemValue wrapping a StructValue with two fields.
	const objectJSON = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:11",
		"value": {
			"@type": "/gno.HeapItemValue",
			"Value": {
				"T": {"@type": "/gno.RefType", "ID": "gno.land/r/demo/x.User"},
				"V": {
					"@type": "/gno.StructValue",
					"Fields": [
						{"T": {"@type": "/gno.PrimitiveType", "value": "16"}, "V": {"@type": "/gno.StringValue", "value": "alice"}},
						{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "HgAAAAAAAAA="}
					]
				}
			}
		}
	}`

	// Type JSON: the resolved StructType with field names.
	const typeJSON = `{
		"typeid": "gno.land/r/demo/x.User",
		"type": {
			"@type": "/gno.StructType",
			"PkgPath": "gno.land/r/demo/x",
			"Fields": [
				{"Name": "Name", "Type": {"@type": "/gno.PrimitiveType", "value": "16"}, "Embedded": false, "Tag": ""},
				{"Name": "Age", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}
			]
		}
	}`

	nodes, err := DecodeObjectJSONWithType([]byte(objectJSON), []byte(typeJSON))
	require.NoError(t, err)
	require.Len(t, nodes, 2, "struct fields surface as top-level page rows")

	assert.Equal(t, "Name", nodes[0].Name, "field 0 labelled from StructType")
	assert.Equal(t, `"alice"`, nodes[0].Value)
	assert.Equal(t, "Age", nodes[1].Name, "field 1 labelled from StructType")
	assert.Equal(t, "30", nodes[1].Value)
}

// TestDecodeObjectJSONWithType_NilTypeFallsBack: a missing/empty type
// response must not break the page — fall back to positional indices.
func TestDecodeObjectJSONWithType_NilTypeFallsBack(t *testing.T) {
	t.Parallel()

	const objectJSON = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:11",
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="}
			]
		}
	}`

	nodes, err := DecodeObjectJSONWithType([]byte(objectJSON), nil)
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	assert.Equal(t, "0", nodes[0].Name, "no type → positional index")
}

// TestDecodeObjectJSON_Func: a stored function (e.g. realm-level Render)
// must render as a single func node — earlier the walker silently produced
// 0 nodes for FuncValue, leading to "no exposed state" placeholder pages.
func TestDecodeObjectJSON_Func(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:9",
		"value": {
			"@type": "/gno.FuncValue",
			"Type": {"@type": "/gno.FuncType",
				"Params": [{"Name": "x", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}],
				"Results": [{"Name": ".res.0", "Type": {"@type": "/gno.PrimitiveType", "value": "32"}, "Embedded": false, "Tag": ""}]},
			"Name": "double",
			"Source": {"@type": "/gno.RefNode",
				"Location": {"PkgPath": "gno.land/r/test", "File": "math.gno",
					"Span": {"Pos": {"Line": "5", "Column": "1"}, "End": {"Line": "7", "Column": "1"}, "Num": "0"}}}
		}
	}`

	nodes, err := DecodeObjectJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1, "stored function must surface as a single node, not 0")
	assert.Equal(t, "func", nodes[0].Kind)
	assert.Equal(t, "double", nodes[0].Name)
	require.NotNil(t, nodes[0].Source, "function with Source location must expose it")
	assert.Equal(t, "math.gno", nodes[0].Source.File)
}

// TestDecodeObjectJSON_Slice: a stored slice with an inline ArrayValue base
// must enumerate elements; one with a RefValue base must emit a single
// expandable handle so the user can drill into the backing array.
func TestDecodeObjectJSON_Slice_Inline(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:5",
		"value": {
			"@type": "/gno.SliceValue",
			"Base": {"@type": "/gno.ArrayValue", "List": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AQAAAAAAAAA="},
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "AgAAAAAAAAA="}
			]},
			"Offset": "0", "Length": "2", "Maxcap": "2"
		}
	}`

	nodes, err := DecodeObjectJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 2)
	assert.Equal(t, "1", nodes[0].Value)
	assert.Equal(t, "2", nodes[1].Value)
}

func TestDecodeObjectJSON_Slice_RefBase(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:5",
		"value": {
			"@type": "/gno.SliceValue",
			"Base": {"@type": "/gno.RefValue", "ObjectID": "ffffffffffffffffffffffffffffffffffffffff:6"},
			"Offset": "0", "Length": "10", "Maxcap": "10"
		}
	}`

	nodes, err := DecodeObjectJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1, "ref-backed slice surfaces as one navigable handle")
	assert.True(t, nodes[0].Expandable)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffff:6", nodes[0].ObjectID)
}

// TestDecodeObjectJSON_Pointer: covers PointerValue at the object root.
// PointerValue with TV inline → expose the target. With a RefValue base →
// expose the navigation handle.
func TestDecodeObjectJSON_Pointer_Inline(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:7",
		"value": {
			"@type": "/gno.PointerValue",
			"TV": {"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "BwAAAAAAAAA="},
			"Index": "0"
		}
	}`

	nodes, err := DecodeObjectJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	assert.Equal(t, "*", nodes[0].Name)
	assert.Equal(t, "7", nodes[0].Value)
}

// TestDecodeObjectJSON_Array covers the qobject_json path for a non-byte
// ArrayValue (List branch). Production hits this when expanding stored arrays.
func TestDecodeObjectJSON_Array(t *testing.T) {
	t.Parallel()

	const fixture = `{
		"objectid": "ffffffffffffffffffffffffffffffffffffffff:9",
		"value": {
			"@type": "/gno.ArrayValue",
			"ObjectInfo": {"ID": "ffffffffffffffffffffffffffffffffffffffff:9"},
			"List": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "CgAAAAAAAAA="},
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "FAAAAAAAAAA="}
			]
		}
	}`

	nodes, err := DecodeObjectJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 2)
	assert.Equal(t, "10", nodes[0].Value)
	assert.Equal(t, "20", nodes[1].Value)
}

// ---- Robustness tests (graceful failure modes) --------------------------

// TestDecodeMalformedJSON ensures the walker errors out (instead of panicking)
// when the backend returns syntactically invalid JSON. The HTTP handler must
// be able to map this to a 5xx without crashing the gnoweb process.
func TestDecodeMalformedJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{"truncated", `{"names": ["x"], "values": [`},
		{"empty body", ``},
		{"not an object", `[1, 2, 3]`},
		{"nonsense", `<<not json>>`},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("walker panicked on malformed input %q: %v", c.name, r)
				}
			}()
			_, err := DecodePkgJSON([]byte(c.raw))
			assert.Error(t, err, "malformed input should produce an error, not a partial result")
		})
	}
}

// TestDecodeEmptyPackage covers a realm with no top-level declarations —
// returns an empty slice (not nil-vs-empty drift, not an error).
// Locks in the contract that handlers can rely on len(nodes) == 0 to render
// an "empty state" view.
func TestDecodeEmptyPackage(t *testing.T) {
	t.Parallel()

	nodes, err := DecodePkgJSON([]byte(`{"names": [], "values": []}`))
	require.NoError(t, err)
	assert.Empty(t, nodes)
}

// ---- Stress / large-render tests ----------------------------------------

// TestDecodeLargeMap_1k checks that decoding a 1000-entry map produces a
// well-formed StateNode tree without panicking and without truncating
// children. Catches regressions where pagination or limits would mask bugs.
func TestDecodeLargeMap_1k(t *testing.T) {
	t.Parallel()

	const n = 1000
	fixture := buildLargeMapFixture(n)

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	m := nodes[0]
	assert.Equal(t, "map", m.Kind)
	require.NotNil(t, m.Length)
	assert.Equal(t, n, *m.Length, "every key must round-trip")
	require.Len(t, m.Children, n, "no children dropped at decode time")

	// Spot-check a few entries.
	assert.Equal(t, `"k0"`, m.Children[0].Name)
	assert.Equal(t, "0", m.Children[0].Value)
	assert.Equal(t, fmt.Sprintf(`"k%d"`, n-1), m.Children[n-1].Name)
}

// TestDecodeDeepStruct exercises 50-deep nesting — a worst-case stack depth
// for the recursive walker. Should not blow up.
func TestDecodeDeepStruct(t *testing.T) {
	t.Parallel()

	const depth = 50
	fixture := buildDeepStructFixture(depth)

	nodes, err := DecodePkgJSON([]byte(fixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)

	// Walk down the tree to confirm depth is preserved.
	cur := nodes[0]
	for i := 0; i < depth; i++ {
		require.Equal(t, "struct", cur.Kind, "level %d is struct", i)
		require.NotEmpty(t, cur.Children, "level %d has children", i)
		cur = cur.Children[0]
	}
	// Bottom: a primitive int.
	assert.Equal(t, "primitive", cur.Kind)
	assert.Equal(t, "1", cur.Value)
}

func BenchmarkDecodeLargeMap_1k(b *testing.B) {
	fixture := []byte(buildLargeMapFixture(1000))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodePkgJSON(fixture); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeDeepStruct_50(b *testing.B) {
	fixture := []byte(buildDeepStructFixture(50))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodePkgJSON(fixture); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeLargeMap_1k_Parallel simulates many concurrent requests
// hitting the same payload — the realistic case when a popular realm is
// browsed by many users at the same time. Validates that the walker has no
// hidden contention (mutex, shared map, …) that would degrade under load.
func BenchmarkDecodeLargeMap_1k_Parallel(b *testing.B) {
	fixture := []byte(buildLargeMapFixture(1000))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := DecodePkgJSON(fixture); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkDecodeMixed_Parallel exercises a realistic mix: each goroutine
// alternates between decoding a small package (top-level summary) and a
// medium one. Closer to a real traffic pattern where users land on a realm
// (top-level) and click into objects (medium).
func BenchmarkDecodeMixed_Parallel(b *testing.B) {
	small := []byte(qpkgFixture)
	medium := []byte(buildLargeMapFixture(200))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			payload := small
			if i%3 == 0 {
				payload = medium
			}
			if _, err := DecodePkgJSON(payload); err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---- Fixture builders ----------------------------------------------------

// buildLargeMapFixture builds a qpkg_json with one map[string]int containing N
// entries: {"k0":0, "k1":1, ...}.
func buildLargeMapFixture(n int) string {
	var b strings.Builder
	b.WriteString(`{"names":["m"],"values":[{"T":{"@type":"/gno.MapType","Key":{"@type":"/gno.PrimitiveType","value":"16"},"Value":{"@type":"/gno.PrimitiveType","value":"32"}},"V":{"@type":"/gno.MapValue","List":{"List":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		// Encode int i as little-endian 8 bytes, base64.
		nbuf := make([]byte, 8)
		for j := 0; j < 8; j++ {
			nbuf[j] = byte(i >> (8 * j))
		}
		nB64 := base64Encode(nbuf)
		fmt.Fprintf(&b,
			`{"Key":{"T":{"@type":"/gno.PrimitiveType","value":"16"},"V":{"@type":"/gno.StringValue","value":"k%d"}},"Value":{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"%s"}}`,
			i, nB64,
		)
	}
	b.WriteString(`]}}}]}`)
	return b.String()
}

// buildDeepStructFixture builds a qpkg_json with one struct nested d levels
// deep, each level containing a single field "x" — the innermost field is an
// int = 1.
func buildDeepStructFixture(d int) string {
	var b strings.Builder
	b.WriteString(`{"names":["root"],"values":[`)
	openLevel := func() {
		b.WriteString(`{"T":{"@type":"/gno.StructType","PkgPath":"test","Fields":[{"Name":"x","Type":{"@type":"/gno.StructType","PkgPath":"test","Fields":[]},"Embedded":false,"Tag":""}]},"V":{"@type":"/gno.StructValue","Fields":[`)
	}
	for i := 0; i < d; i++ {
		openLevel()
	}
	// Innermost: int = 1
	b.WriteString(`{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"AQAAAAAAAAA="}`)
	for i := 0; i < d; i++ {
		b.WriteString(`]}}`)
	}
	b.WriteString(`]}`)
	return b.String()
}

func base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// TestWalkerDepthBound pins the safety cap on recursion. A genuinely
// deep value tree never appears in production realms, but adversarial
// (or buggy) inputs must not blow the renderer's stack — the walker
// yields a "(too deep)" sentinel leaf at maxDecodeDepth instead.
//
// Tests the cap directly via the package-internal entry point. Empty
// TypedValue is fine: depth-check fires BEFORE tv.T == nil branch.
func TestWalkerDepthBound(t *testing.T) {
	t.Parallel()

	// Sanity: a regular int round-trip walks at depth 0 normally.
	const intFixture = `{"names":["x"],"values":[{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"AQAAAAAAAAA="}]}`
	nodes, err := DecodePkgJSON([]byte(intFixture))
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	assert.Equal(t, "1", nodes[0].Value, "shallow walk produces real value")

	// At-cap: walker short-circuits to sentinel regardless of input.
	got := decodeTypedValueAt(maxDecodeDepth, "deep", capturedIntTV(t))
	assert.Equal(t, "(too deep)", got.Type, "at-cap walk yields sentinel")
	assert.Equal(t, "truncated", got.Kind)
	assert.Empty(t, got.Children, "sentinel has no children")

	// One slot below cap: still walks normally — bound is exclusive at depth-=cap.
	below := decodeTypedValueAt(maxDecodeDepth-1, "below", capturedIntTV(t))
	assert.NotEqual(t, "truncated", below.Kind, "depth < cap walks normally")
	assert.Equal(t, "1", below.Value, "depth < cap still decodes value")
}

// capturedIntTV returns a TypedValue for `int(1)` produced by Amino
// JSON unmarshalling — the same shape the walker sees in production.
// Avoids depending on internal gno construction APIs in the test.
func capturedIntTV(t *testing.T) gno.TypedValue {
	t.Helper()
	const wrapper = `{"names":["x"],"values":[{"T":{"@type":"/gno.PrimitiveType","value":"32"},"N":"AQAAAAAAAAA="}]}`
	var resp pkgResponse
	require.NoError(t, amino.UnmarshalJSON([]byte(wrapper), &resp))
	require.Len(t, resp.Values, 1)
	return resp.Values[0]
}
