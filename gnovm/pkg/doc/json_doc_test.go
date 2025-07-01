package doc

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONDocumentation(t *testing.T) {
	dir, err := filepath.Abs("./testdata/integ/hello")
	require.NoError(t, err)
	pkgPath := "gno.land/r/hello"
	expected := &JSONDocumentation{
		PackagePath: "gno.land/r/hello",
		PackageLine: "package hello // import \"hello\"",
		PackageDoc:  "hello is a package for testing\n",
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
			},
			{
				Type:      "",
				Name:      "Inc",
				Signature: "func Inc() int",
				Doc:       "",
				Params:    []*JSONField{},
				Results: []*JSONField{
					{Name: "", Type: "int"},
				},
			},
			{
				Type:      "",
				Name:      "Panic",
				Signature: "func Panic()",
				Doc:       "Panic is a func for testing\n",
				Params:    []*JSONField{},
				Results:   []*JSONField{},
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
			},
		},
		Types: []*JSONType{
			{
				Name:      "myStruct",
				Signature: "type myStruct struct{ a int }",
				Doc:       "myStruct is a struct for testing\n",
			},
		},
	}

	// Get the JSONDocumentation similar to VMKeeper.QueryDoc
	mpkg, err := gnolang.ReadMemPackage(dir, pkgPath)
	require.NoError(t, err)
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	jdoc, err := d.WriteJSONDocumentation()
	require.NoError(t, err)

	assert.Equal(t, expected.JSON(), jdoc.JSON())
}
