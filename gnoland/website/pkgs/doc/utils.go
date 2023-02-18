package doc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"strings"
)

func generateFuncSignature(fn *ast.FuncDecl) string {
	if fn == nil {
		return ""
	}
	var receiver string
	if fn.Recv != nil {
		var receiverNames []string
		for _, field := range fn.Recv.List {
			var fieldName string
			if len(field.Names) > 0 {
				fieldName = field.Names[0].Name
			}
			receiverType := typeString(field.Type)
			receiverNames = append(receiverNames, fmt.Sprintf("%s %s", fieldName, receiverType))
		}
		if len(receiverNames) > 0 {
			receiver = fmt.Sprintf("(%s) ", strings.Join(receiverNames, ", "))
		}
	}
	params := []string{}
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			var paramType string
			if len(param.Names) == 0 {
				// unnamed parameter
				paramType = typeString(param.Type)
			} else {
				// named parameter(s)
				var paramNames []string
				for _, id := range param.Names {
					paramNames = append(paramNames, id.Name)
				}
				paramType = fmt.Sprintf("%s %s", strings.Join(paramNames, ", "), typeString(param.Type))
			}
			params = append(params, paramType)
		}
	}
	results := []string{}
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			var resultType string
			if len(result.Names) == 0 {
				// unnamed result parameter
				resultType = typeString(result.Type)
			} else {
				// named result parameter(s)
				var resultNames []string
				for _, id := range result.Names {
					resultNames = append(resultNames, id.Name)
				}
				resultType = fmt.Sprintf("%s %s", strings.Join(resultNames, ", "), typeString(result.Type))
			}
			results = append(results, resultType)
		}
	}
	var returnType string
	if len(results) == 1 {
		returnType = results[0]
	} else {
		returnType = fmt.Sprintf("(%s)", strings.Join(results, ", "))
	}

	return fmt.Sprintf("func %s%s(%s) %s", receiver, fn.Name.Name, strings.Join(params, ", "), returnType)
}

func extractFunc(x *ast.FuncDecl) *Func {
	fn := Func{
		Doc:       x.Doc.Text(),
		Name:      x.Name.String(),
		Signature: generateFuncSignature(x),
	}
	if x.Recv != nil {
		for _, rcv := range x.Recv.List {
			if ident, ok := rcv.Type.(*ast.Ident); ok {
				fn.Recv = append(fn.Recv, ident.Name)
			}
			if star, ok := rcv.Type.(*ast.StarExpr); ok {
				fn.Recv = append(fn.Recv, types.ExprString(star.X))
			}
		}
	}

	for _, field := range x.Type.Params.List {
		paramNames := []string{}
		for _, name := range field.Names {
			paramNames = append(paramNames, name.Name)
		}
		paramType := typeString(field.Type)
		param := &FuncParam{
			Type:  paramType,
			Names: paramNames,
		}
		fn.Params = append(fn.Params, param)
	}

	if x.Type.Results != nil {
		for _, field := range x.Type.Results.List {
			returnNames := []string{}
			for _, name := range field.Names {
				if name != nil {
					continue
				}
				returnNames = append(returnNames, name.Name)
			}
			returnType := typeString(field.Type)
			ret := &FuncReturn{
				Type:  returnType,
				Names: returnNames,
			}
			fn.Returns = append(fn.Returns, ret)
		}
	}

	return &fn
}

func extractValue(fset *token.FileSet, x *ast.GenDecl) (*Value, error) {
	value := Value{
		Doc: x.Doc.Text(),
	}
	x.Doc = nil
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, x); err != nil {
		return nil, err
	}
	value.Signature = buf.String()
	for _, spec := range x.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for i, name := range valueSpec.Names {
				valueItem := ValueItem{
					Type: typeString(valueSpec.Type),
					Name: name.String(),
				}
				if len(valueSpec.Values) > i {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						valueItem.Value = lit.Value
					} else if ident, ok := valueSpec.Values[i].(*ast.Ident); ok {
						valueItem.Value = ident.Name
					}
				}
				if valueSpec.Doc != nil {
					valueItem.Doc = valueSpec.Doc.Text()
				}
				value.Items = append(value.Items, &valueItem)
			}
		}
	}
	return &value, nil
}

func extractType(fset *token.FileSet, x *ast.TypeSpec) (*Type, error) {
	newType := Type{}
	newType.Name = x.Name.String()
	newType.Doc = x.Doc.Text()

	x.Doc = nil
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, x); err != nil {
		return nil, err
	}

	newType.Definition = buf.String()
	return &newType, nil
}

// This code is inspired by the code at https://cs.opensource.google/go/go/+/refs/tags/go1.20.1:src/go/doc/reader.go;drc=40ed3591829f67e7a116180aec543dd15bfcf5f9;bpv=1;bpt=1;l=124
func typeString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return typeString(t.X)
	case *ast.IndexListExpr:
		return typeString(t.X)
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s.%s", typeString(t.X), t.Sel.Name)
		}
	case *ast.ParenExpr:
		return typeString(t.X)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", typeString(t.X))
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", typeString(t.Elt))
	}
	return ""
}
