package doc

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"go/doc/comment"

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
	mpkg, err := gnolang.ReadMemPackage(dir, pkgPath, gnolang.MPAnyAll)
	require.NoError(t, err)
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	jdoc, err := d.WriteJSONDocumentation()
	require.NoError(t, err)

	assert.Equal(t, expected.JSON(), jdoc.JSON())
}

func TestCreateCustomPrinter(t *testing.T) {
	// Create a test package to get a printer
	dir, err := filepath.Abs("./testdata/integ/hello")
	require.NoError(t, err)
	pkgPath := "gno.land/r/hello"
	mpkg, err := gnolang.ReadMemPackage(dir, pkgPath, gnolang.MPAnyAll)
	require.NoError(t, err)
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	
	opt := &WriteDocumentationOptions{}
	_, pkg, err := d.pkgData.docPackage(opt)
	require.NoError(t, err)

	// Test that createCustomPrinter returns a printer with empty heading IDs
	printer := createCustomPrinter(pkg)
	require.NotNil(t, printer)
	
	// Test that heading ID function returns empty string
	heading := &comment.Heading{Text: []comment.Text{comment.Plain("Test")}}
	id := printer.HeadingID(heading)
	assert.Equal(t, "", id)
}

func TestNormalizedMarkdownPrinter(t *testing.T) {
	// Create a test package to get a printer
	dir, err := filepath.Abs("./testdata/integ/hello")
	require.NoError(t, err)
	pkgPath := "gno.land/r/hello"
	mpkg, err := gnolang.ReadMemPackage(dir, pkgPath, gnolang.MPAnyAll)
	require.NoError(t, err)
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	
	opt := &WriteDocumentationOptions{}
	_, pkg, err := d.pkgData.docPackage(opt)
	require.NoError(t, err)
	printer := createCustomPrinter(pkg)

	// Test with simple comment
	var p comment.Parser
	doc := p.Parse("Simple comment")
	result := normalizedMarkdownPrinter(printer, doc)
	assert.Equal(t, "Simple comment\n", result)

	// Test with backslash escaping
	doc = p.Parse("Comment with \\\\ backslash")
	result = normalizedMarkdownPrinter(printer, doc)
	assert.Equal(t, "Comment with \\\\ backslash\n", result)

	// Test with indented code block
	doc = p.Parse("Comment with:\n\n    code block\n    more code")
	result = normalizedMarkdownPrinter(printer, doc)
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "code block")
	assert.Contains(t, result, "more code")
	assert.Contains(t, result, "```")
}

func TestConvertIndentedCodeBlocksToFenced(t *testing.T) {
	// Test with no code blocks
	input := "Simple markdown\nwith no code"
	result := convertIndentedCodeBlocksToFenced(input)
	assert.Equal(t, input+"\n", result)

	// Test with indented code block
	input = "Text before\n\n    func test() {\n        return true\n    }\n\nText after"
	result = convertIndentedCodeBlocksToFenced(input)
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "func test() {")
	assert.Contains(t, result, "return true")
	assert.Contains(t, result, "```")
	assert.Contains(t, result, "Text before")
	assert.Contains(t, result, "Text after")

	// Test with tab-indented code block
	input = "Text before\n\n\tfunc test() {\n\t\treturn true\n\t}\n\nText after"
	result = convertIndentedCodeBlocksToFenced(input)
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "func test() {")
	assert.Contains(t, result, "return true")
	assert.Contains(t, result, "```")

	// Test with mixed content
	input = "Header\n\n    code1\n    code2\n\nText\n\n    more code\n\nEnd"
	result = convertIndentedCodeBlocksToFenced(input)
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "code1")
	assert.Contains(t, result, "code2")
	assert.Contains(t, result, "```")
	assert.Contains(t, result, "Text")
	assert.Contains(t, result, "more code")
	assert.Contains(t, result, "End")

	// Test with backslash escaping
	input = "Text with \\\\ backslash\n\n    code with \\\\ backslash"
	result = convertIndentedCodeBlocksToFenced(input)
	assert.Contains(t, result, "Text with \\\\ backslash")
	assert.Contains(t, result, "code with \\\\ backslash")
}

func TestNormalizeCodeBlockStream(t *testing.T) {
	// Test with no code blocks
	input := "Simple text\nwith no code"
	reader := strings.NewReader(input)
	var buf bytes.Buffer
	
	err := normalizeCodeBlockStream(reader, &buf)
	require.NoError(t, err)
	result := buf.String()
	assert.Equal(t, "Simple text\nwith no code\n", result)

	// Test with indented code block
	input = "Text before\n\n    func test() {\n        return true\n    }\n\nText after"
	reader = strings.NewReader(input)
	buf.Reset()
	
	err = normalizeCodeBlockStream(reader, &buf)
	require.NoError(t, err)
	result = buf.String()
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "func test() {")
	assert.Contains(t, result, "return true")
	assert.Contains(t, result, "```")

	// Test with tab-indented code block
	input = "Text before\n\n\tfunc test() {\n\t\treturn true\n\t}\n\nText after"
	reader = strings.NewReader(input)
	buf.Reset()
	
	err = normalizeCodeBlockStream(reader, &buf)
	require.NoError(t, err)
	result = buf.String()
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "func test() {")
	assert.Contains(t, result, "return true")
	assert.Contains(t, result, "```")

	// Test with code block at end
	input = "Text before\n\n    code at end"
	reader = strings.NewReader(input)
	buf.Reset()
	
	err = normalizeCodeBlockStream(reader, &buf)
	require.NoError(t, err)
	result = buf.String()
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "code at end")
	assert.Contains(t, result, "```")

	// Test with multiple code blocks
	input = "Text\n\n    code1\n\nText\n\n    code2\n\nEnd"
	reader = strings.NewReader(input)
	buf.Reset()
	
	err = normalizeCodeBlockStream(reader, &buf)
	require.NoError(t, err)
	result = buf.String()
	assert.Contains(t, result, "```go")
	assert.Contains(t, result, "code1")
	assert.Contains(t, result, "```")
	assert.Contains(t, result, "Text")
	assert.Contains(t, result, "code2")
	assert.Contains(t, result, "End")
}

func TestJSONDocumentationWithCodeBlocks(t *testing.T) {
	// Test that JSONDocumentation properly handles code blocks in documentation
	dir, err := filepath.Abs("./testdata/integ/hello")
	require.NoError(t, err)
	pkgPath := "gno.land/r/hello"
	mpkg, err := gnolang.ReadMemPackage(dir, pkgPath, gnolang.MPAnyAll)
	require.NoError(t, err)
	d, err := NewDocumentableFromMemPkg(mpkg, true, "", "")
	require.NoError(t, err)
	
	jdoc, err := d.WriteJSONDocumentation()
	require.NoError(t, err)
	
	// Verify that the JSON contains proper markdown formatting
	jsonStr := jdoc.JSON()
	assert.Contains(t, jsonStr, "package hello")
	assert.Contains(t, jsonStr, "hello is a package for testing")
	
	// Test that functions are properly documented
	for _, fun := range jdoc.Funcs {
		if fun.Name == "Panic" {
			assert.Contains(t, fun.Doc, "Panic is a func for testing")
		}
	}
	
	// Test that types are properly documented
	for _, typ := range jdoc.Types {
		if typ.Name == "myStruct" {
			assert.Contains(t, typ.Doc, "myStruct is a struct for testing")
		}
	}
}








