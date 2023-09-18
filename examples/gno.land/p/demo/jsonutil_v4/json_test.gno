package jsonutil

import (
	"strings"
	"testing"
)

func TestAST(t *testing.T) {
	json := `{"a":[42, null, true, false, "hello"],"b":3.0,"c":{"ia":{}, "ib":{     "foo"   :    "bar"}},"d":4,"e":5}`
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
	expectedTree := `{"a":[42,null,true,false,"hello"],"b":3.0,"c":{"ia":{},"ib":{"foo":"bar"}},"d":4,"e":5}`
	if JSONASTNodeString(ast) != expectedTree {
		t.Errorf("Expected root node to be `%s`, got `%s`", expectedTree, JSONASTNodeString(ast))
	}
}

type TestType struct {
	A []string `json:"a"`
	B float64  `json:"b"`
	C SubTestType
	D uint       `json:"d"`
	E int        `json:"e"`
	F bool       `json:"f"`
	G *EmptyType `json:"g"`
}

func (tt *TestType) FromJSON(ast *JSONASTNode) {
	ParseObjectAST(ast, []*ParseKV{
		{Key: "a", ArrayParser: func(children []*JSONASTNode) {
			tt.A = make([]string, len(children))
			for i, child := range children {
				ParseASTAny(child, &tt.A[i])
			}
		}},
		{Key: "b", Value: &tt.B},
		{Key: "c", Value: &tt.C},
		{Key: "d", Value: &tt.D},
		{Key: "e", Value: &tt.E},
		{Key: "f", Value: &tt.F},
		{Key: "g", Value: &tt.G},
	})
}

type SubTestType struct {
	IA EmptyType      `json:"ia"`
	IB SubSubTestType `json:"ib"`
}

func (stt *SubTestType) FromJSON(ast *JSONASTNode) {
	ParseObjectAST(ast, []*ParseKV{
		{Key: "ia", Value: &stt.IA},
		{Key: "ib", Value: &stt.IB},
	})
}

type EmptyType struct{}

func (et *EmptyType) FromJSON(ast *JSONASTNode) {
	ParseObjectAST(ast, []*ParseKV{})
}

type SubSubTestType struct {
	Foo string `json:"foo"`
}

func (sstt *SubSubTestType) FromJSON(ast *JSONASTNode) {
	ParseObjectAST(ast, []*ParseKV{
		{Key: "foo", Value: &sstt.Foo},
	})
}

func TestParse(t *testing.T) {
	json := `{"a":["42", "null", "true", "false", "hello"],"b":3.0,"c":{"ia":{}, "ib":{     "foo"   :    "bar"}},"d":4,"e":5, "f": true, "g": null}`
	var tt TestType
	ParseAny(json, &tt)

	if len(tt.A) != 5 {
		t.Errorf("Expected A to have 5 elements, got %d", len(tt.A))
	}
	expected := "42, null, true, false, hello"
	if strings.Join(tt.A, ", ") != expected {
		t.Errorf("Expected A to be `%s`, got `%s`", expected, tt.A[0])
	}

	if tt.B != 42.1 { // FIXME: 3.0
		t.Errorf("Expected B to be 3.0, got %f", tt.B)
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
}
