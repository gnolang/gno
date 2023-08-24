package ujson

import (
	"strings"
	"testing"

	"gno.land/p/demo/avl"
)

func TestAST(t *testing.T) {
	json := `{"a":[42, null, true, false, "hello\n\t\r"],"b":-3,"c":{"ia":{}, "ib":{     "foo"   :    "bar"}},"d":4,"e":5}`
	tokens := tokenize(json)
	expected := 44
	if len(tokens) != expected {
		t.Errorf("Expected %d tokens, got %d", expected, len(tokens))
	}
	remainingTokens, ast := parseAST(tokens)
	if len(remainingTokens) != 0 {
		t.Errorf("Expected 0 remaining tokens, got %d", len(remainingTokens))
	}
	if ast.Kind != JSONKindObject {
		t.Errorf("Expected root node to be an object, got %s", ast.Kind)
	}
	expectedTree := `{"a":[42,null,true,false,"hello\n\t\r"],"b":-3,"c":{"ia":{},"ib":{"foo":"bar"}},"d":4,"e":5}`
	if ast.String() != expectedTree {
		t.Errorf("Expected root node to be `%s`, got `%s`", expectedTree, ast.String())
	}
}

type TestType struct {
	A       []string `json:"a"`
	B       int      `json:"b"`
	C       SubTestType
	D       uint       `json:"d"`
	E       int        `json:"e"`
	F       bool       `json:"f"`
	G       *EmptyType `json:"g"`
	AVLTree *avl.Tree  `json:"avlTree"`
}

func (tt *TestType) FromJSON(ast *JSONASTNode) {
	ast.ParseObject([]*ParseKV{
		{Key: "a", ArrayParser: func(children []*JSONASTNode) {
			tt.A = make([]string, len(children))
			for i, child := range children {
				child.ParseAny(&tt.A[i])
			}
		}},
		{Key: "b", Value: &tt.B},
		{Key: "c", Value: &tt.C},
		{Key: "d", Value: &tt.D},
		{Key: "e", Value: &tt.E},
		{Key: "f", Value: &tt.F},
		{Key: "g", Value: &tt.G},
		{Key: "avlTree", CustomParser: func(node *JSONASTNode) {
			t := ""
			tt.AVLTree = node.ParseAVLTree(&t)
		}},
	})
}

func (tt TestType) ToJSON() string {
	return FormatObject([]FormatKV{
		{Key: "a", Value: FormatSlice(tt.A), Raw: true},
		{Key: "b", Value: tt.B},
		{Key: "c", Value: tt.C},
		{Key: "d", Value: tt.D},
		{Key: "e", Value: tt.E},
		{Key: "f", Value: tt.F},
		{Key: "g", Value: tt.G},
		{Key: "avlTree", Value: tt.AVLTree},
	})
}

type SubTestType struct {
	IA EmptyType      `json:"ia"`
	IB SubSubTestType `json:"ib"`
}

func (stt *SubTestType) FromJSON(ast *JSONASTNode) {
	ast.ParseObject([]*ParseKV{
		{Key: "ia", Value: &stt.IA},
		{Key: "ib", Value: &stt.IB},
	})
}

func (stt SubTestType) ToJSON() string {
	return FormatObject([]FormatKV{
		{Key: "ia", Value: stt.IA},
		{Key: "ib", Value: stt.IB},
	})
}

type EmptyType struct{}

func (et *EmptyType) FromJSON(ast *JSONASTNode) {
	ast.ParseObject([]*ParseKV{})
}

func (et EmptyType) ToJSON() string {
	return FormatObject([]FormatKV{})
}

type SubSubTestType struct {
	Foo string `json:"foo"`
}

func (sstt *SubSubTestType) FromJSON(ast *JSONASTNode) {
	ast.ParseObject([]*ParseKV{
		{Key: "foo", Value: &sstt.Foo},
	})
}

func (sstt SubSubTestType) ToJSON() string {
	return FormatObject([]FormatKV{
		{Key: "foo", Value: sstt.Foo},
	})
}

func TestString(t *testing.T) {
	json := `{"a":["42","null","true","false","hello\t\n\r"],"b":-3,"c":{"ia":{},"ib":{"foo":"bar"}},"d":4,"e":5,"f":true,"g":null,"avlTree":{"bar":"foo"}}`
	var tt TestType
	ParseAny(json, &tt)

	if len(tt.A) != 5 {
		t.Errorf("Expected A to have 5 elements, got %d", len(tt.A))
	}
	expected := "42, null, true, false, hello\t\n\r"
	if strings.Join(tt.A, ", ") != expected {
		t.Errorf("Expected A to be `%s`, got `%s`", expected, tt.A[0])
	}

	if tt.B != -3 {
		t.Errorf("Expected B to be -3, got %f", tt.B)
	}

	if tt.D != 4 {
		t.Errorf("Expected D to be 4, got %d", tt.D)
	}

	if tt.E != 5 {
		t.Errorf("Expected E to be 5, got %d", tt.E)
	}

	if !tt.F {
		t.Errorf("Expected F to be true, got false")
	}

	/*
		BUG?: tt.G == <nil> instead of nil
		if tt.G != nil {
			t.Errorf("Expected G to be nil, got %v", tt.G)
		}
	*/

	output := FormatAny(tt)
	expected = `{"a":["42","null","true","false","hello\t\n\r"],"b":-3,"c":{"ia":{},"ib":{"foo":"bar"}},"d":4,"e":5,"f":true,"g":null,"avlTree":{"bar":"foo"}}`
	if output != expected {
		t.Errorf("Expected output to be `%s`, got `%s`", expected, output)
	}
}
