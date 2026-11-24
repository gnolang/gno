package doc

import (
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONDocumentation(t *testing.T) {
	dir, err := filepath.Abs("./testdata/integ/hello")
	require.NoError(t, err)
	pkgPath := "gno.land/r/hello"
	expected := &JSONDocumentation{
		PackagePath: "gno.land/r/hello",
		PackageLine: "package hello // import \"gno.land/r/hello\"",
		PackageDoc:  "hello is a package for testing\n",
		Bugs:        []string{"Bug about myInterface\n"},
		Values: []*JSONValueDecl{
			{
				Signature: "const ConstString = \"const string\"",
				Const:     true,
				Doc:       "",
				Values: []*JSONValue{
					{
						Name: "ConstString",
						Doc:  "",
						Type: "",
					},
				},
				File: "hello.gno",
				Line: 67,
			},
			{
				Signature: "var (\n\tpvString  = \"private string\" // A private var\n\tPubString = \"public string\"\n)",
				Const:     false,
				Doc:       "Test public and private vars\n",
				Values: []*JSONValue{
					{
						Name: "pvString",
						Doc:  "// A private var\n",
						Type: "",
					},
					{
						Name: "PubString",
						Doc:  "",
						Type: "",
					},
				},
				File: "hello.gno",
				Line: 62,
			},
			{
				Signature: "var counter int = 42",
				Const:     false,
				Doc:       "",
				Values: []*JSONValue{
					{
						Name: "counter",
						Doc:  "",
						Type: "int",
					},
				},
				File: "hello.gno",
				Line: 59,
			},
			{
				Signature: "var myStructInst = myStruct{a: 1000}",
				Const:     false,
				Doc:       "",
				Values: []*JSONValue{
					{
						Name: "myStructInst",
						Doc:  "",
						Type: "",
					},
				},
				File: "hello.gno",
				Line: 15,
			},
			{
				Signature: "var sl = []int{1, 2, 3, 4, 5}",
				Const:     false,
				Doc:       "sl is an int array\n",
				Values: []*JSONValue{
					{
						Name: "sl",
						Doc:  "",
						Type: "",
					},
				},
				File: "hello.gno",
				Line: 5,
			},
			{
				Signature: "const myStructConst *myStruct = &myStruct{a: 1000}",
				Const:     true,
				Doc:       "This const belongs to the myStruct type\n",
				Values: []*JSONValue{
					{
						Name: "myStructConst",
						Doc:  "",
						Type: "*myStruct",
					},
				},
				File: "hello.gno",
				Line: 21,
			},
			{
				Signature: "var myStructPtr *myStruct",
				Const:     false,
				Doc:       "This var belongs to the myStruct type\n",
				Values: []*JSONValue{
					{
						Name: "myStructPtr",
						Doc:  "",
						Type: "*myStruct",
					},
				},
				File: "hello.gno",
				Line: 18,
			},
		},
		Funcs: []*JSONFunc{
			{
				Type:      "",
				Name:      "Echo",
				Signature: "func Echo(msg string) (res string)",
				Doc:       "",
				Params: []*JSONField{
					{Name: "msg", Type: "string"},
				},
				Results: []*JSONField{
					{Name: "res", Type: "string"},
				},
				File: "hello.gno",
				Line: 69,
			},
			{
				Type:      "",
				Name:      "GetCounter",
				Signature: "func GetCounter() int",
				Doc:       "",
				Params:    []*JSONField{},
				Results: []*JSONField{
					{Name: "", Type: "int"},
				},
				File: "hello.gno",
				Line: 70,
			},
			{
				Type:      "",
				Name:      "Inc",
				Crossing:  true,
				Signature: "func Inc(cur realm) int",
				Doc:       "",
				Params: []*JSONField{
					{Name: "cur", Type: "realm"},
				},
				Results: []*JSONField{
					{Name: "", Type: "int"},
				},
				File: "hello.gno",
				Line: 71,
			},
			{
				Type:      "",
				Name:      "Panic",
				Signature: "func Panic()",
				Doc:       "Panic is a func for testing\n",
				Params:    []*JSONField{},
				Results:   []*JSONField{},
				File:      "hello.gno",
				Line:      48,
			},
			{
				Type:      "",
				Name:      "fn",
				Signature: "func fn() func(string) string",
				Doc:       "",
				Params:    []*JSONField{},
				Results: []*JSONField{
					{Name: "", Type: "func(string) string"},
				},
				File: "hello.gno",
				Line: 7,
			},
			{
				Type:      "",
				Name:      "pvEcho",
				Signature: "func pvEcho(msg string) string",
				Doc:       "",
				Params: []*JSONField{
					{Name: "msg", Type: "string"},
				},
				Results: []*JSONField{
					{Name: "", Type: "string"},
				},
				File: "hello.gno",
				Line: 72,
			},
			{
				Type:      "myStruct",
				Name:      "Foo",
				Signature: "func (ms myStruct) Foo() string",
				Doc:       "Foo is a method for testing\n",
				Params:    []*JSONField{},
				Results: []*JSONField{
					{Name: "", Type: "string"},
				},
				File: "hello.gno",
				Line: 24,
			},
		},
		Types: []*JSONType{
			{
				Name:  "myAlias",
				Type:  "myStruct",
				Doc:   "Test type aliases\n",
				Alias: true,
				Kind:  "ident",
				File:  "hello.gno",
				Line:  27,
			},
			{
				Name: "myArrayType",
				Type: "[5]int",
				Doc:  "Test array type\n",
				Kind: "array",
				File: "hello.gno",
				Line: 30,
			},
			{
				Name: "myChanType",
				Type: "chan int",
				Doc:  "Test chan type\n",
				Kind: "chan",
				File: "hello.gno",
				Line: 39,
			},
			{
				Name: "myFuncType",
				Type: "func(int) string",
				Doc:  "Test func type\n",
				Kind: "func",
				File: "hello.gno",
				Line: 42,
			},
			{
				Name: "myInterface",
				Type: "interface {\n\terror\n\t// Bar is for testing\n\tBar(x int) string // Bar line comment\n}",
				Doc:  "myInterface is an interface for testing\n",
				Kind: "interface",
				InterElems: []*JSONInterfaceElement{
					{Type: "error"},
					{
						Method: &JSONFunc{
							Type:      "myInterface",
							Name:      "Bar",
							Signature: "Bar(x int) string",
							Doc:       "// Bar is for testing // Bar line comment\n",
							Params: []*JSONField{
								{Name: "x", Type: "int"},
							},
							Results: []*JSONField{
								{Name: "", Type: "string"},
							},
						},
					},
				},
				File: "hello.gno",
				Line: 53,
			},
			{
				Name: "myMapType",
				Type: "map[string]int",
				Doc:  "Test map type\n",
				Kind: "map",
				File: "hello.gno",
				Line: 36,
			},
			{
				Name: "myPointerType",
				Type: "*myStruct",
				Doc:  "Test pointer type\n",
				Kind: "pointer",
				File: "hello.gno",
				Line: 45,
			},
			{
				Name: "mySliceType",
				Type: "[]int",
				Doc:  "Test slice type\n",
				Kind: "slice",
				File: "hello.gno",
				Line: 33,
			},
			{
				Name: "myStruct",
				Type: "struct {\n\t// a is a field\n\ta int // a comment\n}",
				Doc:  "myStruct is a struct for testing\n",
				Kind: "struct",
				Fields: []*JSONField{
					{Name: "a", Type: "int", Doc: "// a is a field\n// a comment\n"},
				},
				File: "hello.gno",
				Line: 10,
			},
		},
	}

	// Get the JSONDocumentation similar to VMKeeper.QueryDoc
	mpkg, err := gnolang.ReadMemPackage(dir, pkgPath, gnolang.MPAnyAll)
	require.NoError(t, err)
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	jdoc, err := d.WriteJSONDocumentation(nil)
	require.NoError(t, err)

	assert.Equal(t, expected.JSON(), jdoc.JSON())
}

func TestJSONDocumentation_Imports(t *testing.T) {
	mpkg := &std.MemPackage{
		Name: "foo",
		Path: "gno.land/p/demo/foo",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package foo\nimport (\n\t\"strings\"\n\t\"gno.land/p/demo/avl\"\n)\n"},
			{Name: "b.gno", Body: "package foo\nimport \"strings\"\n"},      // duplicate import path
			{Name: "a_test.gno", Body: "package foo\nimport \"testing\"\n"}, // test-only, excluded
		},
	}
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	jdoc, err := d.WriteJSONDocumentation(nil)
	require.NoError(t, err)
	require.Equal(t, []string{"gno.land/p/demo/avl", "strings"}, jdoc.Imports)
}

// TestJSONDocumentation_SharedFieldType checks that the names of one field
// declaration share one formatted type: the type can be as long as the source,
// and so can the list of names.
func TestJSONDocumentation_SharedFieldType(t *testing.T) {
	mpkg := &std.MemPackage{
		Name: "foo",
		Path: "gno.land/p/demo/foo",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package foo\n\ntype T struct {\n\tA, B, C map[string]map[string]int\n}\n"},
		},
	}
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	jdoc, err := d.WriteJSONDocumentation(nil)
	require.NoError(t, err)
	fields := jdoc.Types[0].Fields
	require.Len(t, fields, 3)
	require.Equal(t, "map[string]map[string]int", fields[0].Type)
	require.Same(t, unsafe.StringData(fields[0].Type), unsafe.StringData(fields[2].Type))
}

// TestJSONDocumentation_SharedValueDoc checks that the names of one value spec
// share one rendered trailing comment, which can be as long as the source.
func TestJSONDocumentation_SharedValueDoc(t *testing.T) {
	mpkg := &std.MemPackage{
		Name: "foo",
		Path: "gno.land/p/demo/foo",
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package foo\n\nvar A, B, C int // shared doc\n"},
		},
	}
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	jdoc, err := d.WriteJSONDocumentation(nil)
	require.NoError(t, err)
	values := jdoc.Values[0].Values
	require.Len(t, values, 3)
	require.Contains(t, values[0].Doc, "shared doc")
	require.Same(t, unsafe.StringData(values[0].Doc), unsafe.StringData(values[2].Doc))
}
