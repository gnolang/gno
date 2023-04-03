package doc

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateFuncSignature(t *testing.T) {
	testcases := []struct {
		name string
		fn   *ast.FuncDecl
		want string
	}{
		{
			name: "NoParametersNoResults",
			fn:   &ast.FuncDecl{Name: ast.NewIdent("testFunc"), Type: &ast.FuncType{}},
			want: "func testFunc()",
		},
		{
			name: "ParametersNoResults",
			fn: &ast.FuncDecl{
				Name: ast.NewIdent("testFunc"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("param1")},
								Type:  ast.NewIdent("string"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("param2")},
								Type:  ast.NewIdent("int"),
							},
						},
					},
				},
			},
			want: "func testFunc(param1 string, param2 int)",
		},
		{
			name: "NoParametersResults",
			fn: &ast.FuncDecl{
				Name: ast.NewIdent("testFunc"),
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: ast.NewIdent("string"),
							},
							{
								Type: ast.NewIdent("error"),
							},
						},
					},
				},
			},
			want: "func testFunc() (string, error)",
		},
		{
			name: "OneNamedResult",
			fn: &ast.FuncDecl{
				Name: &ast.Ident{
					Name: "testFunc",
				},
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: ast.NewIdent("string"),
								Names: []*ast.Ident{
									{
										Name: "result",
									},
								},
							},
						},
					},
				},
			},
			want: "func testFunc() (result string)",
		},
		{
			name: "TwoNamedResults",
			fn: &ast.FuncDecl{
				Name: &ast.Ident{
					Name: "testFunc",
				},
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: ast.NewIdent("string"),
								Names: []*ast.Ident{
									{
										Name: "result1",
									},
								},
							},
							{
								Type: ast.NewIdent("int"),
								Names: []*ast.Ident{
									{
										Name: "result2",
									},
								},
							},
						},
					},
				},
			},
			want: "func testFunc() (result1 string, result2 int)",
		},
		{
			name: "FunctionParameter",
			fn: &ast.FuncDecl{
				Name: &ast.Ident{Name: "testFunc"},
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.FuncType{
									Params: &ast.FieldList{
										List: []*ast.Field{
											{
												Type: &ast.Ident{Name: "MyType"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: "func testFunc(func(MyType))",
		},
		{
			name: "InterfaceResult",
			fn: &ast.FuncDecl{
				Name: &ast.Ident{
					Name: "testFunc",
				},
				Type: &ast.FuncType{
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.InterfaceType{},
							},
						},
					},
				},
			},
			want: "func testFunc() interface{}",
		},
	}

	for _, c := range testcases {
		assert.Equal(t, c.want, generateFuncSignature(c.fn), c.name)
	}
}

func TestTypeString(t *testing.T) {
	testcases := []struct {
		expr ast.Expr
		want string
	}{
		{&ast.Ident{Name: "int"}, "int"},
		{&ast.Ident{Name: "string"}, "string"},
		{&ast.StarExpr{X: &ast.Ident{Name: "int"}}, "*int"},
		{&ast.ArrayType{Elt: &ast.Ident{Name: "string"}}, "[]string"},
		{&ast.ArrayType{Len: &ast.BasicLit{Value: "5"}, Elt: &ast.Ident{Name: "int"}}, "[5]int"},
		{&ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}}, "map[string]int"},
		{&ast.ChanType{Value: &ast.Ident{Name: "string"}}, "chan string"},
		{&ast.ChanType{Dir: ast.SEND, Value: &ast.Ident{Name: "int"}}, "chan<- int"},
		{&ast.ChanType{Dir: ast.RECV, Value: &ast.Ident{Name: "float64"}}, "<-chan float64"},
		{&ast.StructType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "a"}}, Type: &ast.Ident{Name: "int"}},
					{Names: []*ast.Ident{{Name: "b"}}, Type: &ast.Ident{Name: "string"}},
					{Type: &ast.Ident{Name: "bool"}},
				},
			},
		}, "struct{a int; b string; bool}"},
		{&ast.FuncType{}, "func()"},
		{&ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "x"}}, Type: &ast.Ident{Name: "int"}},
					{Names: []*ast.Ident{{Name: "y"}}, Type: &ast.Ident{Name: "string"}},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.Ident{Name: "float64"}},
				},
			},
		}, "func(x int, y string) float64"},
		{&ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "f"}}, Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{Type: &ast.Ident{Name: "int"}},
								{Type: &ast.Ident{Name: "float64"}},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{},
						},
					}},
				},
			},
			Results: &ast.FieldList{},
		}, "func(f func(int, float64))"},
		{&ast.FuncType{
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.InterfaceType{},
					},
				},
			},
		}, "func() interface{}"},
		{&ast.SliceExpr{X: &ast.Ident{Name: "string"}}, "[]string"},
		{&ast.SliceExpr{X: &ast.Ident{Name: "int"}}, "[]int"},
	}

	for _, c := range testcases {
		assert.Equal(t, c.want, typeString(c.expr))
	}
}

func TestRemovePointer(t *testing.T) {
	testcases := []struct {
		name string
		want string
	}{
		{"MyType", "MyType"},
		{"*MyType", "MyType"},
	}

	for _, c := range testcases {
		assert.Equal(t, c.want, removePointer(c.name))
	}
}

func TestIsFuncExported(t *testing.T) {
	testcases := []struct {
		name string
		fn   *ast.FuncDecl
		want bool
	}{
		{
			name: "exported function without receiver",
			fn: &ast.FuncDecl{
				Name: ast.NewIdent("ExportedFunc"),
			},
			want: true,
		},
		{
			name: "unexported function without receiver",
			fn: &ast.FuncDecl{
				Name: ast.NewIdent("unexportedFunc"),
			},
			want: false,
		},
		{
			name: "exported method with exported pointer receiver",
			fn: &ast.FuncDecl{
				Name: ast.NewIdent("ExportedMethod"),
				Recv: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{ast.NewIdent("p")},
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: "MyType",
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "exported method with unexported pointer receiver",
			fn: &ast.FuncDecl{
				Name: ast.NewIdent("ExportedMethod"),
				Recv: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{ast.NewIdent("p")},
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: "myType",
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}

	for _, c := range testcases {
		assert.Equal(t, c.want, isFuncExported(c.fn), c.name)
	}
}
